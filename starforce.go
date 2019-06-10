package main

import (
  "regexp"
  "log"
  "gopkg.in/yaml.v2"
  "io/ioutil"
  "bytes"
  "math/big"
  "net/http"
  "net/url"
  "html/template"
  "strconv"
  "github.com/leekchan/accounting"
)

type gearLevel struct {
  Level int `yaml:"level"`
  Costs [15]int64 `yaml:costs"`
}

type probabilities struct {
  Success [15]int64 `yaml:"success"`
  Fail [15]int64 `yaml:"fail"`
  Destroy [15]int64 `yaml:"destroy"`
}

type formData struct {
  Level int
  //Safeguards [5]bool
  //rCost int64 // replacement cost
  //Events [4]bool
  Start int
  End int
}

type Page struct {
  Amount string
  Level string
  Start string
  End string
  Event1 string
  Event2 string
  Event3 string
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
var nextCosts map[int]map[int]*big.Rat
var validPath *regexp.Regexp

func init() {
  m = readGearLevels()
  pr = readProbabilities()
  nextCosts = make(map[int]map[int]*big.Rat)
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

func nextStarCost(equipLevel int, level int) *big.Rat {
  costs := m[equipLevel]
  i := level - 10
  fail := pr.Fail[i]
  boom := pr.Destroy[i]
  success := pr.Success[i]

  if _, ok := nextCosts[equipLevel]; !ok {
    nextCosts[equipLevel] = make(map[int]*big.Rat)
  }


  if val, ok := nextCosts[equipLevel][level]; ok {
    return val
  }

  cost := costs[i]
  pCost := func() int64 { if level<11 {return 0} else {return costs[i-1]} }()
  ppCost := func() int64 { if level<12 {return 0} else {return costs[i-2]} }()
  expAtmpts := big.NewRat(1000, success)
  expCost := new(big.Rat)
  expCost.Mul(big.NewRat(cost, 1), expAtmpts)

  if level == 10 {
    nextCosts[equipLevel][level] = expCost
    return expCost
  }

  failRatio := big.NewRat(fail, success)

  if level == 11 {
    expCost.Add(expCost, failRatio.Mul(failRatio, nextStarCost(equipLevel, 10)))
    nextCosts[equipLevel][level] = expCost
    return expCost
  }

  dblFailRatio := big.NewRat(fail*pr.Fail[level-11], success*1000)
  dblFailCost := new(big.Rat)

  if level == 12 {
    dblFailCost.Add(big.NewRat(ppCost, 1), nextStarCost(equipLevel, level - 1))
    dblFailCost.Mul(dblFailRatio, dblFailCost)
    failRatio.Mul(failRatio, big.NewRat(pCost, 1))
    expCost.Add(expCost, dblFailCost)
    expCost.Add(expCost, failRatio)
    // TODO: add replacement cost
    nextCosts[equipLevel][level] = expCost
    return expCost
  }

  expPrevCost := new(big.Rat)
  expPrevCost.Set(nextStarCost(equipLevel, level - 1))
  dblFailCost.Mul(dblFailRatio, big.NewRat(ppCost, 1))

  boomAtmpts := big.NewRat(boom, 1000)
  boomAtmpts.Mul(boomAtmpts, expAtmpts)

  if level == 13 {
    noSuccess := big.NewRat(1000 - pr.Success[i-1], 1000)
    noSuccess.Mul(failRatio, noSuccess)
    noSuccess.Add(noSuccess, boomAtmpts)
    expPrevCost.Mul(expPrevCost, noSuccess)
    failRatio.Mul(failRatio, big.NewRat(pCost, 1))
    expCost.Add(expCost, dblFailCost)
    expCost.Add(expCost, failRatio)
    expCost.Add(expCost, expPrevCost)
    // TODO: add replacement cost
    nextCosts[equipLevel][level] = expCost
    return expCost
  }

  returnCost := big.NewRat(0, 1)
  for i := 12; i < level; i++ {
    returnCost.Add(returnCost, nextStarCost(equipLevel, i))
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
    nextCosts[equipLevel][level] = expCost
    return expCost
  }

  if level == 15 || level == 20 {
    returnCost.Mul(returnCost, big.NewRat(boom, 1000))
    returnCost.Mul(returnCost, big.NewRat(1000, success))
    expCost.Add(returnCost, expCost)
    nextCosts[equipLevel][level] = expCost
    return expCost
  }

  if level == 16 || level == 21 {
    returnCost.Mul(returnCost, boomAtmpts)
    failRatio.Mul(failRatio, nextStarCost(equipLevel, level - 1))
    expCost.Add(expCost, returnCost)
    expCost.Add(expCost, failRatio)
    nextCosts[equipLevel][level] = expCost
    return expCost
  }

  return expCost
}

func overallCosts(input formData) string {
  res := big.NewRat(0, 1)
  for i := input.Start; i < input.End; i++ {
    res.Add(res, nextStarCost(input.Level, i))
  }
  ac := accounting.Accounting{Symbol: "", Precision: 2}
  return ac.FormatMoney(res)
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

func handler(w http.ResponseWriter, r *http.Request) {
  x := validPath.FindStringSubmatch(r.URL.Path)
  if x == nil {
    http.NotFound(w, r)
    return
  }
  r.ParseForm()
  p := Page{"", "150", "10", "17", "off", "off", "off"}
  if len(r.Form) > 0 {
    start, _ := strconv.Atoi(r.Form["start"][0])
    end, _ := strconv.Atoi(r.Form["end"][0])
    p.Start = r.Form["start"][0]
    p.End = r.Form["end"][0]
    if end <= start {
      p.Amount = "Invalid input"
    } else {
      level, _ := strconv.Atoi(r.Form["level"][0])
      input := formData{level, start, end}
      p.Amount = "Expected cost is " + overallCosts(input)
      p.Level = r.Form["level"][0]
      checkEvents(&p, r.Form)
    }
  }
  templates := template.Must(template.ParseFiles("template/starforce.html"))
  templates.ExecuteTemplate(w, "starforce.html", p)
}

func main() {
  http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
  http.HandleFunc("/", handler)
  log.Fatal(http.ListenAndServe(":8080", nil))
}

