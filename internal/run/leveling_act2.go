package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
	"github.com/hectorgimenez/koolo/internal/reader"
	"github.com/hectorgimenez/koolo/internal/ui"
)

func (a Leveling) act2() action.Action {
	running := false
	return action.NewFactory(func(d data.Data) action.Action {
		if running || d.PlayerUnit.Area != area.LutGholein {
			return nil
		}

		running = true
		// Find Horadric Cube
		_, found := d.Items.Find("HoradricCube", item.LocationInventory, item.LocationStash)
		if found {
			a.logger.Info("Horadric Cube found, skipping quest")
		} else {
			a.logger.Info("Horadric Cube not found, starting quest")
			return a.findHoradricCube()
		}

		quests := a.builder.GetCompletedQuests(2)
		if quests[4] {
			return a.duriel(quests[1])
		}

		_, horadricStaffFound := d.Items.Find("HoradricStaff", item.LocationInventory, item.LocationStash, item.LocationEquipped)

		// Find Staff of Kings
		_, found = d.Items.Find("StaffOfKings", item.LocationInventory, item.LocationStash, item.LocationEquipped)
		if found || horadricStaffFound {
			a.logger.Info("StaffOfKings found, skipping quest")
		} else {
			a.logger.Info("StaffOfKings not found, starting quest")
			return a.findStaff()
		}

		// Find Amulet
		_, found = d.Items.Find("AmuletOfTheViper", item.LocationInventory, item.LocationStash, item.LocationEquipped)
		if found || horadricStaffFound {
			a.logger.Info("Amulet of the Viper found, skipping quest")
		} else {
			a.logger.Info("Amulet of the Viper not found, starting quest")
			return a.findAmulet()
		}

		// Summoner
		a.logger.Info("Starting summoner quest")
		return a.summoner()
	})
}

func (a Leveling) radament() action.Action {
	return action.NewChain(func(d data.Data) (actions []action.Action) {
		actions = append(actions,
			a.builder.WayPoint(area.SewersLevel2Act2),
			a.builder.MoveToArea(area.SewersLevel3Act2),
		)

		// TODO: Find Radament (use 355 object to locate him)
		return
	})
}

func (a Leveling) findHoradricCube() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		return []action.Action{
			a.builder.WayPoint(area.HallsOfTheDeadLevel2),
			a.builder.MoveToArea(area.HallsOfTheDeadLevel3),
			a.builder.MoveTo(func(d data.Data) (data.Position, bool) {
				a.logger.Info("Horadric Cube chest found, moving to that room")
				chest, found := d.Objects.FindOne(object.HoradricCubeChest)

				return chest.Position, found
			}),
			a.builder.ClearAreaAroundPlayer(15),
			action.BuildStatic(func(d data.Data) []step.Step {
				a.logger.Info("Opening Horadric cube chest...")
				return []step.Step{
					step.InteractObject(object.HoradricCubeChest, func(d data.Data) bool {
						chest, _ := d.Objects.FindOne(object.HoradricCubeChest)
						return !chest.Selectable
					}),
				}
			}),
			a.builder.ItemPickup(true, 10),
		}
	})
}

func (a Leveling) findStaff() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		return []action.Action{
			a.builder.WayPoint(area.FarOasis),
			a.builder.MoveToArea(area.MaggotLairLevel1),
			a.builder.MoveToArea(area.MaggotLairLevel2),
			a.builder.MoveToArea(area.MaggotLairLevel3),
			a.builder.MoveTo(func(d data.Data) (data.Position, bool) {
				a.logger.Info("Staff Of Kings chest found, moving to that room")
				chest, found := d.Objects.FindOne(object.StaffOfKingsChest)

				return chest.Position, found
			}),
			a.builder.ClearAreaAroundPlayer(15),
			action.BuildStatic(func(d data.Data) []step.Step {
				a.logger.Info("Opening Staff Of Kings chest...")
				return []step.Step{
					step.InteractObject(object.StaffOfKingsChest, func(d data.Data) bool {
						chest, _ := d.Objects.FindOne(object.StaffOfKingsChest)
						return !chest.Selectable
					}),
				}
			}),
			a.builder.ItemPickup(true, 10),
		}
	})
}

func (a Leveling) findAmulet() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		return []action.Action{
			a.builder.WayPoint(area.LostCity),
			a.builder.MoveToArea(area.ValleyOfSnakes),
			a.builder.MoveToArea(area.ClawViperTempleLevel1),
			a.builder.MoveToArea(area.ClawViperTempleLevel2),
			a.builder.MoveTo(func(d data.Data) (data.Position, bool) {
				a.logger.Info("Altar found, moving closer")
				chest, found := d.Objects.FindOne(object.TaintedSunAltar)

				return chest.Position, found
			}),
			action.BuildStatic(func(d data.Data) []step.Step {
				a.logger.Info("Opening altar...")
				return []step.Step{
					step.InteractObject(object.TaintedSunAltar, func(d data.Data) bool {
						chest, _ := d.Objects.FindOne(object.TaintedSunAltar)
						return !chest.Selectable
					}),
				}
			}),
			a.builder.ItemPickup(true, 10),
		}
	})
}

func (a Leveling) summoner() action.Action {
	return action.NewChain(func(d data.Data) (actions []action.Action) {
		actions = append(actions, Summoner{baseRun: a.baseRun}.BuildActions()...)

		// Try to use the portal and discover the waypoint
		actions = append(actions,
			action.BuildStatic(func(d data.Data) []step.Step {
				return []step.Step{
					step.InteractObject(object.YetAnotherTome, func(d data.Data) bool {
						_, found := d.Objects.FindOne(object.PermanentTownPortal)
						return found
					}),
				}
			}),

			action.BuildStatic(func(d data.Data) []step.Step {
				return []step.Step{
					step.InteractObject(object.PermanentTownPortal, func(d data.Data) bool {
						return d.PlayerUnit.Area == area.CanyonOfTheMagi
					}),
				}
			}),

			a.builder.DiscoverWaypoint(),
			a.builder.ReturnTown(),
			action.BuildStatic(func(d data.Data) []step.Step {
				return []step.Step{
					step.InteractNPC(npc.Atma),
					step.SyncStep(func(d data.Data) error {
						hid.PressKey("esc")
						return nil
					}),
				}
			}),
		)

		return
	})
}

func (a Leveling) prepareStaff() action.Action {
	return action.NewChain(func(d data.Data) (actions []action.Action) {
		horadricStaff, found := d.Items.Find("HoradricStaff", item.LocationInventory, item.LocationStash, item.LocationEquipped)
		if found {
			a.logger.Info("Horadric Staff found!")
			if horadricStaff.Location == item.LocationStash {
				a.logger.Info("It's in the stash, let's pickup it (not done yet)")

				return []action.Action{action.BuildStatic(func(d data.Data) []step.Step {
					return []step.Step{
						step.InteractObject(object.Bank, func(d data.Data) bool {
							return d.OpenMenus.Stash
						}),
						step.SyncStep(func(d data.Data) error {
							screenPos := ui.GetScreenCoordsForItem(horadricStaff)
							hid.MovePointer(screenPos.X, screenPos.Y)

							hid.KeyDown("control")
							helper.Sleep(300)
							hid.Click(hid.LeftButton)
							helper.Sleep(200)
							hid.KeyUp("control")
							helper.Sleep(300)
							hid.PressKey("esc")
							return nil
						}),
					}
				})}
			}

			return nil
		}

		staff, found := d.Items.Find("StaffOfKings", item.LocationInventory, item.LocationStash, item.LocationEquipped)
		if !found {
			a.logger.Info("Staff of Kings not found, skipping")
			return nil
		}

		amulet, found := d.Items.Find("AmuletOfTheViper", item.LocationInventory, item.LocationStash, item.LocationEquipped)
		if !found {
			a.logger.Info("AmuletOfTheViper not found, skipping")
			return nil
		}

		return []action.Action{
			a.builder.CubeAddItems(staff, amulet),
			a.builder.CubeTransmute(),
		}
	})
}

func (a Leveling) duriel(staffAlreadyUsed bool) action.Action {
	a.logger.Info("Starting Duriel....")
	tombs := []area.Area{area.TalRashasTomb1, area.TalRashasTomb2, area.TalRashasTomb3, area.TalRashasTomb4, area.TalRashasTomb5, area.TalRashasTomb6, area.TalRashasTomb7}

	var realTomb area.Area
	for _, tomb := range tombs {
		_, _, objects, _ := reader.CachedMapData.NPCsExitsAndObjects(data.Position{}, tomb)
		for _, obj := range objects {
			if obj.Name == object.HoradricOrifice {
				realTomb = tomb
				break
			}
		}
	}

	if realTomb == 0 {
		a.logger.Info("Could not find the real tomb :(")
		return nil
	}

	return action.NewChain(func(d data.Data) (actions []action.Action) {
		if !staffAlreadyUsed {
			actions = append(actions, a.prepareStaff())
		}

		// Move close to the Horadric Orifice
		actions = append(actions,
			a.builder.WayPoint(area.CanyonOfTheMagi),
			a.char.Buff(),
			a.builder.MoveToArea(realTomb),
			a.builder.MoveTo(func(d data.Data) (data.Position, bool) {
				orifice, _ := d.Objects.FindOne(object.HoradricOrifice)

				return orifice.Position, true
			}),
		)

		// If staff has not been used, then put it in the orifice and wait for the entrance to open
		if !staffAlreadyUsed {
			actions = append(actions,
				a.builder.ClearAreaAroundPlayer(20),
				action.BuildStatic(func(d data.Data) []step.Step {
					return []step.Step{
						step.InteractObject(object.HoradricOrifice, func(d data.Data) bool {
							return d.OpenMenus.Anvil
						}),
						step.SyncStep(func(d data.Data) error {
							staff, _ := d.Items.Find("HoradricStaff", item.LocationInventory)

							screenPos := ui.GetScreenCoordsForItem(staff)
							hid.MovePointer(screenPos.X, screenPos.Y)

							helper.Sleep(300)
							hid.Click(hid.LeftButton)
							hid.MovePointer(ui.AnvilCenterX, ui.AnvilCenterY)
							helper.Sleep(300)
							hid.Click(hid.LeftButton)
							helper.Sleep(300)
							hid.MovePointer(ui.AnvilBtnX, ui.AnvilBtnY)
							helper.Sleep(500)
							hid.Click(hid.LeftButton)
							helper.Sleep(20000)

							return nil
						}),
					}
				}),
			)
		}

		potsToBuy := 4
		if d.MercHPPercent() > 0 {
			potsToBuy = 8
		}

		// Return to the city, ensure we have pots and everything, and get some thawing potions
		actions = append(actions,
			a.builder.ReturnTown(),
			a.builder.VendorRefill(),
			a.builder.BuyAtVendor(npc.Lysander, action.VendorItemRequest{
				Item:     "ThawingPotion",
				Quantity: potsToBuy,
				Tab:      4,
			}),
			action.BuildStatic(func(d data.Data) []step.Step {
				return []step.Step{
					step.SyncStep(func(d data.Data) error {
						hid.PressKey(config.Config.Bindings.OpenInventory)
						x := 0
						for _, itm := range d.Items.ByLocation(item.LocationInventory) {
							if itm.Name != "ThawingPotion" {
								continue
							}

							pos := ui.GetScreenCoordsForItem(itm)
							hid.MovePointer(pos.X, pos.Y)
							helper.Sleep(500)

							if x > 3 {
								hid.Click(hid.LeftButton)
								helper.Sleep(300)
								hid.MovePointer(ui.MercAvatarPositionX, ui.MercAvatarPositionY)
								helper.Sleep(300)
								hid.Click(hid.LeftButton)
							} else {
								hid.Click(hid.RightButton)
							}
							x++
						}

						hid.PressKey("esc")
						return nil
					}),
				}
			}),
			a.builder.UsePortalInTown(),
			a.char.Buff(),
		)

		return append(actions,
			action.BuildStatic(func(d data.Data) []step.Step {
				_, found := d.Objects.FindOne(object.DurielsLairPortal)
				if found {
					return []step.Step{
						step.InteractObject(object.DurielsLairPortal, func(d data.Data) bool {
							return d.PlayerUnit.Area == area.DurielsLair
						}),
					}
				}

				return nil
			}),
			a.char.KillDuriel(),
			a.builder.ItemPickup(true, 30),
			a.builder.MoveToCoords(data.Position{
				X: 22577,
				Y: 15613,
			}),
			action.BuildStatic(func(d data.Data) []step.Step {
				return []step.Step{
					step.InteractNPCWithCheck(npc.Tyrael, func(d data.Data) bool {
						obj, found := d.Objects.FindOne(object.TownPortal)
						if found && pather.DistanceFromMe(d, obj.Position) < 10 {
							return true
						}

						return false
					}),
				}
			}, action.CanBeSkipped()),
			a.builder.ReturnTown(),
			a.builder.MoveToCoords(data.Position{
				X: 5092,
				Y: 5144,
			}),
			action.BuildStatic(func(d data.Data) []step.Step {
				return []step.Step{
					step.InteractNPC(npc.Jerhyn),
					step.KeySequence("esc"),
				}
			}),
			a.builder.MoveToCoords(data.Position{
				X: 5195,
				Y: 5060,
			}),
			action.BuildStatic(func(d data.Data) []step.Step {
				return []step.Step{
					step.InteractNPC(npc.Meshif),
					step.KeySequence("home", "down", "enter"),
				}
			}),
		)
	})
}
