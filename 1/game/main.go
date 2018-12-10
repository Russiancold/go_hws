package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

type player struct {
	position    room
	hasBackpack bool
	backpack    backpack
}

type room struct {
	name             string
	interactiveItems map[string][]InteractiveItem //мапа - стринга с расположением (на столе, на стуле...) на айтемы
	doors            []door
	hint             string
	lookAroundString string
}

func (r room) getItemsString() string {
	var str string
	for key, val := range r.interactiveItems {
		str += key + ": "
		for _, v := range val {
			str += v.getName() + ", "
		}
	}
	if len(str) > 2 {
		return str[:len(str)-2]
	} else {
		return ""
	}
}

func (r *room) getItem(arg string) InteractiveItem {
	for i, val := range r.interactiveItems {
		for _, item := range val {
			if item.getName() == arg {
				var idxToDelete int
				for idx := range val {
					if val[idx].getName() == arg {
						r.interactiveItems[i] = append(val[:idxToDelete], val[idxToDelete+1:]...)
						if len(r.interactiveItems[i]) == 0 {
							delete(r.interactiveItems, i)
						}
					}
				}
				return item
			}
		}
	}
	return nil
}

func (r room) getConnectedRoomsString() string {
	str := "можно пройти - "
	rooms := gameMap[r.name]
	if rooms == nil || len(rooms) == 0 {
		return "выхода нет"
	} else {
		for _, val := range rooms {
			str += val.name + ", "
		}
		return str[:len(str)-2]
	}
}

func (r room) getHint() string {
	return r.hint
}

func addItem(r *room, location string, newItems []InteractiveItem) {
	if _, err := r.interactiveItems[location]; !err {
		r.interactiveItems[location] = newItems
	} else {
		r.interactiveItems[location] = append(r.interactiveItems[location], newItems...)
	}
}

func addConnectedRooms(keyRoom room, connectedRooms []room) {
	_, exists := gameMap[keyRoom.name]
	if !exists {
		gameMap[keyRoom.name] = connectedRooms
	} else {
		gameMap[keyRoom.name] = append(gameMap[keyRoom.name], connectedRooms...)
	}
}

var gameMap map[string][]room = map[string][]room{}
var back backpack
var doors []door
var pawn player

func main() {
	initGame()
	reader := bufio.NewReader(os.Stdin)
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		} else {
			fmt.Println(handleCommand(text))
		}
	}

}

func initGame() {
	gameMap = map[string][]room{}
	back = backpack{}
	doors = nil
	pawn = player{}

	//initialize items
	var hallKey key
	var book uselessItem
	var tea uselessItem
	var d door
	back.name = "рюкзак"
	hallKey.name = "ключи"
	book.name = "конспекты"
	tea.name = "чай"

	//initialize rooms
	kitchen := room{"кухня", map[string][]InteractiveItem{
		"на столе": {tea},
	}, []door{}, "надо собрать рюкзак и идти в универ", "ты находишься на кухне"}
	hall := room{"коридор", map[string][]InteractiveItem{}, []door{d}, "", "ничего интересного"}
	bedroom := room{"комната", map[string][]InteractiveItem{
		"на столе": {hallKey, book},
		"на стуле": {back},
	}, []door{}, "", "ты в своей комнате"}
	street := room{"улица", map[string][]InteractiveItem{}, []door{d}, "", "на улице весна"}

	//connect this s**t
	addConnectedRooms(kitchen, []room{hall})
	addConnectedRooms(hall, []room{kitchen, bedroom, street})
	addConnectedRooms(bedroom, []room{hall})
	addConnectedRooms(street, []room{hall})

	//set items
	d.setRoom1(hall)
	d.setRoom2(street)
	d.setKey(hallKey)
	doors = append(doors, d)

	pawn = player{kitchen, false, back}
}

func handleCommand(command string) string {
	command, args := strings.Fields(command)[0], strings.Fields(command)[1:]
	switch command {
	case "идти":
		if len(args) > 1 {
			return "Too many arguments"
		}
		pass, err := changeRoomByName(pawn.position, args[0])
		if err == nil {
			if pass {
				return pawn.position.lookAroundString + ", " + pawn.position.getConnectedRoomsString()
			} else {
				return "дверь закрыта"
			}
		} else {
			return "нет пути в " + args[0]
		}
	case "осмотреться":
		return constructLookAroundString(pawn.position.lookAroundString, pawn.position.getItemsString(), pawn.position.getHint(), pawn.position.getConnectedRoomsString())
	case "взять":
		if len(args) > 1 {
			return "Too many arguments"
		}
		if !pawn.hasBackpack {
			return "некуда класть"
		} else {
			item := pawn.position.getItem(args[0])
			if item == nil {
				return "нет такого"
			}
			if item.getName() == "рюкзак" {
				pawn.hasBackpack = true
				return "вы надели: рюкзак"
			}
			back.stash = append(back.stash, item)
			return "предмет добавлен в инвентарь: " + item.getName()
		}
	case "надеть":
		if len(args) > 1 {
			return "Too many arguments"
		}
		item := pawn.position.getItem(args[0])
		if item == nil {
			return "нет такого"
		}
		if item.getName() == "рюкзак" {
			pawn.hasBackpack = true
			return "вы надели: рюкзак"
		} else {
			return "как ты это наденешь?"
		}
	case "применить":
		for _, val := range back.stash {
			if val.getName() == args[0] {
				if args[0] == "ключи" {
					if args[1] != "дверь" {
						return "не к чему применить" //hardcoded
					}
					if doors[0].r1.name == pawn.position.name || doors[0].r2.name == pawn.position.name {
						doors[0].openClose(true)
						return "дверь открыта"
					} else {
						return "нет дверей"
					}
				}
				return "нельзя применить" + args[0]
			}
		}
		return "нет предмета в инвентаре - " + args[0]
	default:
		return "неизвестная команда"
	}
	return "fucked up("
}

func constructLookAroundString(la string, items string, hint string, rooms string) string {
	if items == "" {
		if hint == "" {
			return la + ", " + rooms
		}
		return la + ", " + hint + ", " + rooms
	} else {
		if hint == "" {
			return la + ", " + items + ", " + rooms
		}
		return la + ", " + items + ", " + hint + ", " + rooms
	}
}

//true - room exists, no door/ door is open; false - room with door
func changeRoomByName(r room, name string) (bool, error) {
	for idx := range gameMap[r.name] {
		if gameMap[r.name][idx].name == name {
			for i := range doors {
				if (doors[i].r1.name == name && doors[i].r2.name == pawn.position.name) ||
					(doors[i].r1.name == pawn.position.name && doors[i].r2.name == name) {
					if doors[i].isOpen {
						pawn.position = gameMap[r.name][idx]
						return true, nil
					} else {
						return false, nil
					}
				}
			}
			pawn.position = gameMap[r.name][idx]
			return true, nil
		}
	}
	return false, errors.New("no such room")
}
