package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/leekchan/accounting"
	"gopkg.in/yaml.v2"
)

type gearLevel struct {
	Level int       `yaml:"level"`
	Costs [15]int64 `yaml:costs"`
}

type probabilities struct {
	Success [15]int64 `yaml:"success"`
	Fail    [15]int64 `yaml:"fail"`
	Destroy [15]int64 `yaml:"destroy"`
}

type formData struct {
	Level int
	//Safeguards [5]bool
	rCost int64 // replacement cost
	//Events [4]bool
	Start int
	End   int
}

type Page struct {
	Amount          string
	Level           string
	Start           string
	End             string
	Event1          string
	Event2          string
	Event3          string
	Safeguards1     string
	Safeguards2     string
	Safeguards3     string
	Safeguards4     string
	Safeguards5     string
	ReplacementCost string
}

func readGearLevels() map[int][15]int64 {
	yamlFile, err := ioutil.ReadFile("levels.yml")
	if err != nil {
		panic(err)
	}
	r := bytes.NewReader(yamlFile)
	dec := yaml.NewDecoder(r)
	var g gearLevel
	var res map[int][15]int64
	res = make(map[int][15]int64)
	for dec.Decode(&g) == nil {
		//fmt.Printf("Level : %v\nCosts : %v\n", g.Level, g.Costs)
		res[g.Level] = g.Costs
	}
	return res
}

func readProbabilities() probabilities {
	yamlFile, err := ioutil.ReadFile("probabilities.yml")
	if err != nil {
		panic(err)
	}
	var p probabilities
	err = yaml.Unmarshal(yamlFile, &p)
	if err != nil {
		panic(err)
	}
	return p
}

var m map[int][15]int64
var pr probabilities
var nextCosts map[int]*big.Rat
var validPath *regexp.Regexp

func init() {
	m = readGearLevels()
	pr = readProbabilities()
	nextCosts = make(map[int]*big.Rat)
	validPath = regexp.MustCompile("^/$")
}

func isGeneralCase(level int) bool {
	switch level {
	case
		14,
		17,
		18,
		19,
		22,
		23,
		24:
		return true
	}
	return false
}

func nextStarCost(equipLevel int, rCost int64, level int) *big.Rat {
	costs := m[equipLevel]
	i := level - 10
	fail := pr.Fail[i]
	boom := pr.Destroy[i]
	success := pr.Success[i]

	if val, ok := nextCosts[level]; ok {
		return val
	}

	cost := costs[i]
	pCost := func() int64 {
		if level < 11 {
			return 0
		} else {
			return costs[i-1]
		}
	}()
	ppCost := func() int64 {
		if level < 12 {
			return 0
		} else {
			return costs[i-2]
		}
	}()
	expAtmpts := big.NewRat(1000, success)
	expCost := new(big.Rat)
	expCost.Mul(big.NewRat(cost, 1), expAtmpts)

	replacement := new(big.Rat)

	if level == 10 {
		nextCosts[level] = expCost
		return expCost
	}

	failRatio := big.NewRat(fail, success)

	if level == 11 {
		temp := nextStarCost(equipLevel, rCost, 10)
		expCost.Add(expCost, failRatio.Mul(failRatio, temp))
		nextCosts[level] = expCost
		return expCost
	}

	dblFailRatio := big.NewRat(fail*pr.Fail[level-11], success*1000)
	dblFailCost := new(big.Rat)

	if level == 12 {
		temp := nextStarCost(equipLevel, rCost, level-1)
		dblFailCost.Add(big.NewRat(ppCost, 1), temp)
		dblFailCost.Mul(dblFailRatio, dblFailCost)
		failRatio.Mul(failRatio, big.NewRat(pCost, 1))
		expCost.Add(expCost, dblFailCost)
		expCost.Add(expCost, failRatio)
		replacement.Mul(big.NewRat(boom, success), big.NewRat(rCost, 1))
		expCost.Add(expCost, replacement)
		nextCosts[level] = expCost
		return expCost
	}

	expPrevCost := new(big.Rat)
	expPrevCost.Set(nextStarCost(equipLevel, rCost, level-1))
	dblFailCost.Mul(dblFailRatio, big.NewRat(ppCost, 1))

	boomAtmpts := big.NewRat(boom, 1000)
	boomAtmpts.Mul(boomAtmpts, expAtmpts)

	replacement.Mul(failRatio, big.NewRat(pr.Destroy[i-1], 1000))
	replacement.Add(replacement, big.NewRat(boom, success))
	replacement.Mul(big.NewRat(rCost, 1), replacement)

	if level == 13 {
		noSuccess := big.NewRat(1000-pr.Success[i-1], 1000)
		noSuccess.Mul(failRatio, noSuccess)
		noSuccess.Add(noSuccess, boomAtmpts)
		expPrevCost.Mul(expPrevCost, noSuccess)
		failRatio.Mul(failRatio, big.NewRat(pCost, 1))
		expCost.Add(expCost, dblFailCost)
		expCost.Add(expCost, failRatio)
		expCost.Add(expCost, expPrevCost)
		expCost.Add(expCost, replacement)
		nextCosts[level] = expCost
		return expCost
	}

	returnCost := big.NewRat(0, 1)
	for i := 12; i < level; i++ {
		returnCost.Add(returnCost, nextStarCost(equipLevel, rCost, i))
	}

	if isGeneralCase(level) {
		dblFailCost2 := new(big.Rat)
		dblFailCost2.Mul(dblFailRatio, expPrevCost)
		fCost := new(big.Rat)
		fCost.Mul(failRatio, big.NewRat(pCost, 1))
		failRatio.Mul(failRatio, big.NewRat(pr.Destroy[i-1], 1000))
		boomAtmpts.Add(failRatio, boomAtmpts)
		returnCost.Mul(boomAtmpts, returnCost)
		expCost.Add(expCost, fCost)
		expCost.Add(expCost, dblFailCost)
		expCost.Add(expCost, dblFailCost2)
		expCost.Add(expCost, returnCost)
		expCost.Add(expCost, replacement)
		nextCosts[level] = expCost
		return expCost
	}

	if level == 15 || level == 20 {
		replacement.Mul(big.NewRat(boom, success), big.NewRat(rCost, 1))
		returnCost.Mul(returnCost, big.NewRat(boom, 1000))
		returnCost.Mul(returnCost, big.NewRat(1000, success))
		expCost.Add(returnCost, expCost)
		expCost.Add(replacement, expCost)
		nextCosts[level] = expCost
		return expCost
	}

	if level == 16 || level == 21 {
		replacement.Mul(big.NewRat(boom, success), big.NewRat(rCost, 1))
		returnCost.Mul(returnCost, boomAtmpts)
		failRatio.Mul(failRatio, nextStarCost(equipLevel, rCost, level-1))
		expCost.Add(expCost, returnCost)
		expCost.Add(expCost, failRatio)
		expCost.Add(expCost, replacement)
		nextCosts[level] = expCost
		return expCost
	}

	return expCost
}

func overallCosts(input formData) string {
	res := big.NewRat(0, 1)
	nextCosts = make(map[int]*big.Rat)
	for i := input.Start; i < input.End; i++ {
		res.Add(res, nextStarCost(input.Level, input.rCost, i))
	}
	ac := accounting.Accounting{Symbol: "", Precision: 2}
	return ac.FormatMoney(res)
}

// NewPage inits page
func NewPage() Page {
	p := Page{}
	p.Level = "150"
	p.Start = "10"
	p.End = "17"
	p.Event1 = ""
	p.Event2 = ""
	p.Event3 = ""
	p.Safeguards1 = ""
	p.Safeguards2 = ""
	p.Safeguards3 = ""
	p.Safeguards4 = ""
	p.Safeguards5 = ""
	p.ReplacementCost = "0"
	return p
}

func checkEvents(p *Page, form url.Values) {
	if val, ok := form["event1"]; ok {
		p.Event1 = val[0]
	}
	if val, ok := form["event2"]; ok {
		p.Event2 = val[0]
	}
	if val, ok := form["event3"]; ok {
		p.Event3 = val[0]
	}
}

func checkReplacementCost(p *Page, form url.Values) {
	if val, ok := form["rcost"]; ok {
		p.ReplacementCost = val[0]
	}
}

func checkSafeguards(p *Page, form url.Values) {
	if val, ok := form["sg12"]; ok {
		p.Safeguards1 = val[0]
	}
	if val, ok := form["sg13"]; ok {
		p.Safeguards2 = val[0]
	}
	if val, ok := form["sg14"]; ok {
		p.Safeguards3 = val[0]
	}
	if val, ok := form["sg15"]; ok {
		p.Safeguards4 = val[0]
	}
	if val, ok := form["sg16"]; ok {
		p.Safeguards5 = val[0]
	}
}

func checkForm(p *Page, form url.Values) {
	checkEvents(p, form)
	checkSafeguards(p, form)
	checkReplacementCost(p, form)
}

func handler(w http.ResponseWriter, r *http.Request) {
	x := validPath.FindStringSubmatch(r.URL.Path)
	if x == nil {
		http.NotFound(w, r)
		return
	}
	r.ParseForm()
	p := NewPage()
	if len(r.Form) > 0 {
		start, _ := strconv.Atoi(r.Form["start"][0])
		end, _ := strconv.Atoi(r.Form["end"][0])
		p.Start = r.Form["start"][0]
		p.End = r.Form["end"][0]
		if end <= start {
			p.Amount = "End star level must be greater than start star level"
		} else {
			level, _ := strconv.Atoi(r.Form["level"][0])
			p.Level = r.Form["level"][0]
			checkForm(&p, r.Form)
			if b, _ := regexp.MatchString(`^[0-9]+`, r.Form["rcost"][0]); !b {
				p.Amount = "Please enter a non-negative, numeric replacement cost"
			} else {
				rcost, _ := strconv.ParseInt(r.Form["rcost"][0], 10, 64)
				input := formData{level, rcost, start, end}
				p.Amount = "Expected cost is " + overallCosts(input) + " mesos"
			}
		}
	}
	templates := template.Must(template.ParseFiles("template/starforce.html"))
	templates.ExecuteTemplate(w, "starforce.html", p)
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", handler)
	fmt.Println("Listening...")
	fmt.Println("Available on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
