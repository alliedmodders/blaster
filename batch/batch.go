// vim: set ts=4 sw=4 tw=99 noet:
//
// Blaster (C) Copyright 2014 AlliedModders LLC
// Licensed under the GNU General Public License, version 3 or higher.
// See LICENSE.txt for more details.
package batch

// A batch is a list of arbitrary items.
type Batch interface {
	Item(index int) interface{}
	Len() int
}

// Callback function to process items.
type Callback func(item interface{})

// A batch processor feeds items into a goroutine for processing.
type BatchProcessor struct {
	callback Callback
	maxTasks int

	batchQueue     chan Batch
	stopCommand    chan bool
	finishedSignal chan bool
	taskDone       chan bool
	stopped        bool

	// These are only modified from the process goroutine.
	worklist    []interface{} // Pending items to create tasks for.
	outstanding int           // Number of remaining tasks we're waiting on.
}

// Create a new batch processor.
func NewBatchProcessor(callback Callback, maxTasks int) *BatchProcessor {
	processor := &BatchProcessor{
		callback: callback,
		maxTasks: maxTasks,

		// We expect batch communication to be fast, so we let it block. The
		// master takes time to reply anyway.
		//
		// Note: we rely on this being synchronous in that sending "stop" to
		// the process routine could die before a batch is pulled out of the
		// queue.
		batchQueue: make(chan Batch),

		// Neither of these should be synchronous.
		stopCommand:    make(chan bool, 1),
		finishedSignal: make(chan bool, 1),

		// Notifications from completed processors. Buffer size doesn't really
		// matter but we'd rather not block to push.
		taskDone: make(chan bool, maxTasks),
	}

	go processor.waitForBatches()

	return processor
}

// Adds a batch to the batch processor.
func (this *BatchProcessor) AddBatch(batch Batch) {
	this.batchQueue <- batch
}

// Signals that no more batches are incoming, and then waits for batch
// processing to complete.
func (this *BatchProcessor) Finish() {
	this.send_stop(false)
}

// Forcefully terminates batch processing. This only shuts down the worker
// routine. Individual processing tasks will continue.
func (this *BatchProcessor) Terminate() {
	this.send_stop(true)
}

func (this *BatchProcessor) send_stop(terminate bool) {
	// Don't re-enter this function.
	if this.stopped {
		return
	}
	this.stopped = true

	// Signal to the processing routine that it should stop.
	this.stopCommand <- terminate

	// Wait for it to return.
	<-this.finishedSignal
}

// This must only be invoked from enqueueBatch() or waitForBatches().
func (this *BatchProcessor) enqueueItem(item interface{}) {
	this.outstanding++

	// Avoid entraining local state by passing everything through the closure.
	go (func(callback Callback, taskDone chan bool, item interface{}) {
		defer (func() {
			taskDone <- true
		})()

		callback(item)
	})(this.callback, this.taskDone, item)
}

// This must only be invoked from waitForBatches(). It enqueues tasks available
// in a batch.
func (this *BatchProcessor) enqueueBatch(batch Batch) {
	index := 0

	// Enqueue everything into goroutines.
	for this.outstanding < this.maxTasks && index < batch.Len() {
		this.enqueueItem(batch.Item(index))
		index++
	}

	// Add any remaining items to the worklist.
	for i := index; i < batch.Len(); i++ {
		this.worklist = append(this.worklist, batch.Item(i))
	}
}

// This should only be called from processBatch().
func (this *BatchProcessor) workRemaining() bool {
	return len(this.worklist) > 0 || this.outstanding > 0
}

// This runs in its own goroutine.
func (this *BatchProcessor) waitForBatches() {
	// Setup local state.
	stopped := false
	terminated := false

	for {
		select {
		case batch := <-this.batchQueue:
			this.enqueueBatch(batch)

		case <-this.taskDone:
			// A single task has completed.
			this.outstanding--

			if len(this.worklist) > 0 {
				// Pop an item off the worklist. This is unreachable after
				// Terminate().
				item := this.worklist[len(this.worklist)-1]
				this.worklist = this.worklist[:len(this.worklist)-1]

				this.enqueueItem(item)
				continue
			}

			if !this.workRemaining() && stopped {
				// If there's no work left to do, and the parent thread is
				// waiting on us to finish, then leave now.
				if !terminated {
					this.finishedSignal <- true
				}
				return
			}

		case terminated = <-this.stopCommand:
			stopped = true

			if terminated {
				// Detach the worklist so we don't enqueue anything else. We
				// do notify the parent thread early, since it has no reason
				// to wait on us.
				this.worklist = nil
				this.finishedSignal <- true

				// If outstanding is 0, we can exit. Otherwise, there's a
				// possible deadlock: tasks could keep pumping into taskDone,
				// but if nothing is consuming it, those tasks will linger
				// around. In that case, we leave this task running until
				// all tasks have finished.
				//
				// This isn't strictly necessary since the buffering size of
				// the task channel == max number of tasks, but we do it
				// anyway for safety.
				if this.outstanding == 0 {
					return
				}
			}

			if !this.workRemaining() {
				this.finishedSignal <- true
				return
			}
		}
	}
}
