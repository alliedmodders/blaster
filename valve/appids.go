// vim: set ts=4 sw=4 tw=99 noet:
//
// Blaster (C) Copyright 2014 AlliedModders LLC
// Licensed under the GNU General Public License, version 3 or higher.
// See LICENSE.txt for more details.
package valve

type AppId int32

const (
	App_Unknown AppId = 0

	// HL1
	App_CS       AppId = 10
	App_TFC      AppId = 20
	App_DOD      AppId = 30
	App_DMC      AppId = 40
	App_OP4      AppId = 50
	App_Ricochet AppId = 60
	App_HL       AppId = 70
	App_CS_CZ    AppId = 80

	// HL2
	App_SDK2006          AppId = 215
	App_SDK2007          AppId = 218
	App_CSS              AppId = 240
	App_DODS             AppId = 300
	App_HL2DM            AppId = 320
	App_HLDMS            AppId = 360
	App_TF2              AppId = 440
	App_L4D1             AppId = 500
	App_L4D2             AppId = 550
	App_AlienSwarm       AppId = 630
	App_CSGO             AppId = 730
	App_DarkMessiah      AppId = 2130
	App_TheShip          AppId = 2400
	App_BloodyGoodTime   AppId = 2450
	App_GarrysMod        AppId = 4000
	App_ZombiePanic      AppId = 17500
	App_AgeOfChivalry    AppId = 17510
	App_Synergy          AppId = 17520
	App_DIPRIP           AppId = 17530
	App_EternalSilence   AppId = 17550
	App_PVK              AppId = 17570
	App_Dystopia         AppId = 17580
	App_InsurgencyMod    AppId = 17700
	App_NuclearDawn      AppId = 17710
	App_Smashball        AppId = 17730
	App_EmpiresMod       AppId = 17740
	App_DinoDDay         AppId = 70000
	App_EYE              AppId = 91700
	App_Insurgency       AppId = 222880
	App_NoMoreRoomInHell AppId = 224260
	App_BladeSymphony    AppId = 225600
	App_Contagion        AppId = 238430
	App_SDK2013          AppId = 243750
	App_Neotokyo         AppId = 244630
	App_FortressForever  AppId = 253530
	App_FistfulOfFrags   AppId = 265630
	App_ModularCombat    AppId = 349480
	App_CodenameCURE     AppId = 355180
	App_BlackMesa        AppId = 362890
	App_DayOfInfamy      AppId = 447820
	App_IOSoccer         AppId = 673560
)

var HL1Apps = []AppId{
	App_CS,
	App_TFC,
	App_DOD,
	App_DMC,
	App_OP4,
	App_Ricochet,
	App_HL,
	App_CS_CZ,
}

var HL2Apps = []AppId{
	App_SDK2006,
	App_SDK2007,
	App_CSS,
	App_DODS,
	App_HL2DM,
	App_HLDMS,
	App_TF2,
	App_L4D1,
	App_L4D2,
	App_AlienSwarm,
	App_CSGO,
	App_DarkMessiah,
	App_TheShip,
	App_BloodyGoodTime,
	App_GarrysMod,
	App_ZombiePanic,
	App_AgeOfChivalry,
	App_Synergy,
	App_DIPRIP,
	App_EternalSilence,
	App_PVK,
	App_Dystopia,
	App_InsurgencyMod,
	App_NuclearDawn,
	App_Smashball,
	App_EmpiresMod,
	App_DinoDDay,
	App_EYE,
	App_Insurgency,
	App_NoMoreRoomInHell,
	App_BladeSymphony,
	App_Contagion,
	App_SDK2013,
	App_Neotokyo,
	App_FortressForever,
	App_FistfulOfFrags,
	App_ModularCombat,
	App_CodenameCURE,
	App_BlackMesa,
	App_DayOfInfamy,
	App_IOSoccer,
}

func IsPreOrangeBoxApp(appId AppId) bool {
	switch appId {
	case App_SDK2006, App_EternalSilence, App_InsurgencyMod, App_Neotokyo, App_FortressForever:
		return true
	default:
		return false
	}
}
