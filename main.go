package main

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

type player struct {
	name string
	lvl  int
	hp,
	maxhp int
	strength,
	dexterity,
	vitality,
	intellect int
	gold   int
	weapon string
	battle bool
	hunted map[string]bool
}

type item struct {
	dmgMin int
	dmgMax int
	price  int
}

type battle struct {
	active bool // default 'false'
	mName  string
	mHp    int
}

type monster struct {
	dmgMin int
	dmgMax int
	hp     int
}

var turn = 1

var monsters = map[string]monster{
	"rat":    {1, 2, 15}, // dmgMin, dmgMax, hp
	"jackal": {2, 3, 30},
	"goblin": {2, 5, 100},
	"troll":  {5, 10, 250},
	"dragon": {12, 18, 500},
}

var thisBattle battle

var items = map[string]item{
	"nothing":     {dmgMin: 1, dmgMax: 2, price: 0},
	"dagger":      {dmgMin: 1, dmgMax: 4, price: 10},
	"short sword": {dmgMin: 2, dmgMax: 4, price: 50},
	"long sword":  {dmgMin: 2, dmgMax: 6, price: 150},
	"battle axe":  {dmgMin: 3, dmgMax: 8, price: 250},
	"halberd":     {dmgMin: 1, dmgMax: 12, price: 500},
}

var p1 = &player{
	name:      "",
	lvl:       1,
	hp:        10,
	maxhp:     10,
	strength:  1,
	dexterity: 1,
	vitality:  1,
	intellect: 1,
	gold:      10,
	weapon:    "nothing",
	battle:    false,
	hunted:    map[string]bool{"none": false},
}

var p2 = &player{
	name:      "",
	lvl:       1,
	hp:        10,
	maxhp:     10,
	strength:  1,
	dexterity: 1,
	vitality:  1,
	intellect: 1,
	gold:      10,
	weapon:    "nothing",
	battle:    false,
	hunted:    map[string]bool{"none": false},
}

func gameloop(r *bufio.Reader) {
	cmd := ""

	for cmd != "exit" {
		rand.Seed(time.Now().UTC().UnixNano())
		var thisP, otherP *player
		var err error

		p1.maxhp = p1.vitality*3 + 7
		p2.maxhp = p2.vitality*3 + 7

		//thisP = p1 ... to test single chracter

		if turn%2 == 0 {
			thisP = p1
			otherP = p2
		} else {
			thisP = p2
			otherP = p1
		}

		fmt.Printf("%s's turn #%d // HP: %d    Lvl: %d    Gold: %d ---> ", thisP.name, turn, thisP.hp, thisP.lvl, thisP.gold)

		if thisP.battle { // if current player already in the battle
			err = pve(thisP)
		} else { // regular situation, current player not in the battle

			// get input
			line, _ := r.ReadString('\n')
			line = strings.TrimSpace(line)

			tokens := strings.Split(line, " ")

			cmd = tokens[0]
			args := tokens[1:]

			switch cmd {
			case "pve":
				err = pve(thisP)
			case "duel":
				err = duel(thisP, otherP)
			case "heal":
				err = heal(thisP)
			case "job":
				err = job(thisP)
			case "train":
				if len(args) > 0 {
					err = train(thisP, args[0])
				} else {
					err = errors.New("train <stat> (eg str, dex, vit, int)")
				}
			case "buy":
				if len(args) > 0 {
					err = buy(thisP, args...)
				} else {
					err = errors.New("buy <item> (eg dagger, sword)")
				}
			case "stats":
				err = stats(thisP)
			default: // idle
				err = errors.New("wrong command. Try again")
			}
		}

		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("%s turn ended with HP: %d    Lvl: %d    Gold: %d \n", thisP.name, thisP.hp, thisP.lvl, thisP.gold)
			fmt.Println("====================")
			// regen a bit if not in a battle
			if !thisP.battle && thisP.hp != 0 && thisP.hp < thisP.maxhp {
				thisP.hp += thisP.lvl
			}
			if thisP.hp > thisP.maxhp {
				thisP.hp = thisP.maxhp
			}
			// next turn
			turn++
		}
	}
}

func main() {
	fmt.Println(`                ~~=~~~ Welcome to Hotseat-text-RPG ~~~=~~-

	It's a game for two players who can use one keyboard (or you
	can play alone with two characters) and perform actions by turn.
	Goal: kill the Dragon with as fewer turns as possible.

	Stats:
	Str (strength)  - more damage
	Dex (dexterity) - dodge chance
	Vit (vitality)  - more HPs
	Int (intellect) - moar gold

	Each unique defeated monster increase player's level which increases
	the chance to meet Dragon and gives boni to healing.

	List of actions:
	pve, duel, heal, job, train <stat>, buy <item>, stats
	`)
	r := bufio.NewReader(os.Stdin)

	// player names
	fmt.Print("Please enter the name of 1st character: ")
	p1.name, _ = r.ReadString('\n')
	p1.name = strings.TrimSpace(p1.name)
	fmt.Print("And the name of 2nd character: ")
	p2.name, _ = r.ReadString('\n')
	p2.name = strings.TrimSpace(p2.name)

	gameloop(r)
}

func pve(p *player) error {
	if p.hp < 1 {
		return errors.New("you need to heal first")
	}

	if !thisBattle.active { // init battle

		mon_lvl := (turn + 10) / 20
		if rand.Intn(2) == 0 {
			mon_lvl = rand.Intn(p.lvl)
		}

		switch mon_lvl {
		case 0:
			thisBattle.mName = "rat"
		case 1:
			thisBattle.mName = "jackal"
		case 2:
			thisBattle.mName = "goblin"
		case 3:
			thisBattle.mName = "troll"
		default:
			thisBattle.mName = "dragon"
		}
		thisBattle.active = true
		thisBattle.mHp = monsters[thisBattle.mName].hp
		p.battle = true
	} else if thisBattle.active && !p.battle { // second player enter battle
		p.battle = true
	}

	// p turn
	weapon, ok := items[p.weapon]
	if !ok {
		panic("inv state. no p weapon")
	}
	dice := rand.Intn(weapon.dmgMax-weapon.dmgMin) + weapon.dmgMin + p.strength
	fmt.Printf("%s hit %s for %d damage.\n", p.name, thisBattle.mName, dice)
	thisBattle.mHp -= dice

	// m turn
	m := monsters[thisBattle.mName]
	// if only 1 player fights - then monster attack each turn
	// But if we have 2 players, m attack each 2nd turn
	if thisBattle.mHp != 0 && ((!p1.battle || !p2.battle) || turn%2 == 0) { // using p1/p2 globals

		dodge := 20 - p.dexterity
		if dodge < 2 {
			dodge = 2
		}

		if rand.Intn(dodge) != 0 {
			dice := rand.Intn(m.dmgMax-m.dmgMin) + m.dmgMin
			fmt.Printf("~~~ %s hit %s for %d damage.\n", strings.Title(thisBattle.mName), p.name, dice)
			p.hp -= dice
		} else {
			fmt.Printf("%s dodged %s's attack!\n", p.name, thisBattle.mName)
		}
	}

	// p defeated -> kick from battle
	if p.hp < 1 {
		fmt.Printf("%s become unconscious and leave the battle..\n", p.name)
		p.battle = false
		if p.hp < 0 {
			p.hp = 0
			if p.lvl > 1 {
				fmt.Printf("%s lost one level.\n", p.name)
				p.lvl -= 1
			}
		}
	}

	// both p defeated -> reset monster
	if !p1.battle && !p2.battle {
		thisBattle.active = false
	}

	// m defeated
	if thisBattle.mHp < 1 {
		reward := rand.Intn(m.hp) + 5

		// if it's 1st time when we beat this kind - add mob to 'hunter' list
		if p1.battle && p2.battle { // if both currently in fight
			if _, ok := p1.hunted[thisBattle.mName]; !ok {
				p1.hunted[thisBattle.mName] = true
			}
			if _, ok := p2.hunted[thisBattle.mName]; !ok {
				p2.hunted[thisBattle.mName] = true
			}
			p1.lvl++
			p2.lvl++
			p1.gold += reward / 2
			p2.gold += reward / 2
			fmt.Printf("%s and %s won the battle and looted %d coins each!\n", p1.name, p2.name, reward/2)
		} else { // if alone
			if _, ok := p.hunted[thisBattle.mName]; !ok {
				p.hunted[thisBattle.mName] = true
			}
			p.lvl++
			p.gold += reward
			fmt.Printf("%s won the battle and looted %d coins!\n", p.name, reward)
		}

		// reset mobs
		p1.battle, p2.battle = false, false // globals
		thisBattle.active = false

		// Win condition
		if thisBattle.mName == "dragon" {
			fmt.Printf("You won! It took %d turns to defeat the Dragon! ;)\n", turn)
			fmt.Println("Press 'q' to quit the game.")
			bufio.NewReader(os.Stdin).ReadBytes('q')
			os.Exit(0)
		}
	}

	return nil
}

func heal(p *player) error {

	heal := p.maxhp / 5 * p.lvl
	if p.hp < p.maxhp {
		p.hp += heal
	}
	if p.hp > p.maxhp {
		p.hp = p.maxhp
	}
	fmt.Printf("You managed to heal %d damage.\n", heal)
	return nil
}

func duel(att *player, def *player) error {

	if att.hp != att.maxhp || def.hp != def.maxhp {
		return fmt.Errorf("can't duel while hurt")
	}

	if def.battle {
		return fmt.Errorf("can't duel with other player during PvE battle")
	}

	for att.hp > 0 && def.hp > 0 {
		if rand.Intn(2) == 0 { // attacker turn
			weapon, ok := items[att.weapon]
			if !ok {
				panic("no attacker weapon, report bug to devs")
			}
			dice := rand.Intn(weapon.dmgMax-weapon.dmgMin) + weapon.dmgMin
			fmt.Printf("%s hit %s for %d damage.\n", att.name, def.name, dice)
			def.hp -= dice
		} else { // defender turn
			weapon, ok := items[def.weapon]
			if !ok {
				panic("no defender weapon, report bug to devs")
			}
			dice := rand.Intn(weapon.dmgMax-weapon.dmgMin) + weapon.dmgMin
			fmt.Printf("%s hit %s for %d damage.\n", def.name, att.name, dice)
			att.hp -= dice
		}
	}
	if att.hp > 0 {
		fmt.Printf("%s won and got %s's money... %d coins!\n", att.name, def.name, def.gold)
		att.gold += def.gold
		def.gold = 0
	} else {
		fmt.Printf("%s won and got %s's money... %d coins!\n", def.name, att.name, att.gold)
		def.gold += att.gold
		att.gold = 0
	}

	// heal after pvp
	att.hp = att.maxhp
	def.hp = def.maxhp

	return nil
}

func job(p *player) error {
	gold := rand.Intn(p.strength+p.dexterity+p.vitality) + (p.intellect * 2)

	jobs := []string{
		"You dug up some ore",
		"You chopped some wood",
		"You catch some fish",
		"You tinkered some goods",
		"You found a small treasure",
		"You harvested some crops",
		"You mined some coal",
		"You gathered some berries",
		"You carved some stone",
		"You brewed some potions",
		"You looted a chest",
	}

	fmt.Printf("%s and sold it for %d gold.\n", jobs[rand.Intn(len(jobs))], gold)
	p.gold += gold
	return nil
}

func train(p *player, stat string) error {
	statsSum := p.strength + p.dexterity + p.vitality + p.intellect
	if p.gold < statsSum*2 {
		return fmt.Errorf(
			"you have only %d gold, while to train you need %d", p.gold, statsSum*2)
	}

	switch stat {
	case "str":
		p.strength++
	case "dex":
		p.dexterity++
	case "vit":
		p.vitality++
	case "int":
		p.intellect++
	default:
		return errors.New("which stat to train? Str, dex, vit or int?")
	}

	p.gold -= statsSum * 2

	return nil
}

func buy(p *player, itemSlice ...string) error {
	item := ""
	for i, v := range itemSlice {
		if len(itemSlice) != i+1 { // last element
			item += v + " " // to add space (short_sword)
		} else {
			item += v
		}
	}

	itemBuy, ok := items[item]
	if !ok || item == "nothing" {
		return fmt.Errorf(`there is no such item. Choose one of this (item:price):
		dagger: 10; short sword:50, long sword: 150; battle axe: 250; halberd: 500`)
	}

	if p.gold < itemBuy.price {
		return fmt.Errorf("you need %d gold to buy a %s, while you have %d gold", itemBuy.price, item, p.gold)
	}

	itemBuy.price -= p.intellect - 1 // not a bug, but feature ;) bargain!
	if itemBuy.price < 1 {
		itemBuy.price = 1
	}

	p.weapon = item
	p.gold -= itemBuy.price
	fmt.Println("Great! You bought", item)
	return nil
}

func stats(p *player) error {
	fmt.Println("Level:", p.lvl)
	fmt.Println("MaxHP:", p.maxhp)
	fmt.Println("HP:", p.hp)
	fmt.Println("strength:", p.strength)
	fmt.Println("Dexterity:", p.dexterity)
	fmt.Println("Vitality:", p.vitality)
	fmt.Println("Intellect:", p.intellect)
	fmt.Println("Gold:", p.gold)
	fmt.Printf("Hunted monsters: ")
	for m := range p.hunted {
		if m != "none" {
			fmt.Printf("%v ", m)
		}
	}

	for k := range items {
		if k == p.weapon {
			return fmt.Errorf("\nCurrent weapon: %s", k)
		}
	}
	return errors.New("you don't have any items")
}
