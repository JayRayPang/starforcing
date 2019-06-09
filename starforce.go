package main

import (
  "log"
  "gopkg.in/yaml.v2"
  "io/ioutil"
  "bytes"
  "math/big"
  "net/http"
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

var m map[int][15]int64 = readGearLevels()
var pr probabilities = readProbabilities()

func expectedBooms(pr probabilities, level int) *big.Rat {
  fail := pr.Fail[level - 10]
  boom := pr.Destroy[level - 10]
  success := pr.Success[level - 10]

  expAtmpts := new(big.Rat)
  expAtmpts.SetFrac(big.NewInt(1000), big.NewInt(success))

  expFails := new(big.Rat)
  expFails.Sub(expAtmpts, big.NewRat(1, 1))

  expBoom := new(big.Rat)
  expBoom.SetFrac(big.NewInt(boom), big.NewInt(boom + fail))
  expBoom.Mul(expBoom, expFails)

  downgrade := new(big.Rat)
  downgrade.SetFrac(big.NewInt(fail), big.NewInt(boom + fail))

  prevBooms := big.NewRat(0, 1)

  for i := 12; i < level; i++ {
    prevBooms.Add(prevBooms, expectedBooms(pr, i))
  }

  prevBooms.Mul(downgrade, prevBooms)
  return expBoom.Add(expBoom, prevBooms)
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

func nextStarCost(costs [15]int64, level int) *big.Rat {
  i := level - 10
  fail := pr.Fail[i]
  boom := pr.Destroy[i]
  success := pr.Success[i]

  cost := costs[i]
  pCost := func() int64 { if level<11 {return 0} else {return costs[i-1]} }()
  ppCost := func() int64 { if level<12 {return 0} else {return costs[i-2]} }()
  expAtmpts := big.NewRat(1000, success)
  expCost := new(big.Rat)
  expCost.Mul(big.NewRat(cost, 1), expAtmpts)

  if level == 10 {
    return expCost
  }

  failRatio := big.NewRat(fail, success)

  if level == 11 {
    expCost.Add(expCost, failRatio.Mul(failRatio, nextStarCost(costs, 10)))
    return expCost
  }

  dblFailRatio := big.NewRat(fail*pr.Fail[level-11], success*1000)
  dblFailCost := new(big.Rat)

  if level == 12 {
    dblFailCost.Add(big.NewRat(ppCost, 1), nextStarCost(costs, level - 1))
    dblFailCost.Mul(dblFailRatio, dblFailCost)
    failRatio.Mul(failRatio, big.NewRat(pCost, 1))
    expCost.Add(expCost, dblFailCost)
    expCost.Add(expCost, failRatio)
    // TODO: add replacement cost
    return expCost
  }

  expPrevCost := nextStarCost(costs, level - 1)
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
    return expCost
  }

  returnCost := big.NewRat(0, 1)
  for i := 12; i < level; i++ {
    returnCost.Add(returnCost, nextStarCost(costs, i))
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
    return expCost
  }

  if level == 15 || level == 20 {
    returnCost.Mul(returnCost, big.NewRat(boom, 1000))
    returnCost.Mul(returnCost, big.NewRat(1000, success))
    expCost.Add(returnCost, expCost)
    return expCost
  }

  if level == 16 || level == 21 {
    returnCost.Mul(returnCost, boomAtmpts)
    failRatio.Mul(failRatio, nextStarCost(costs, level - 1))
    expCost.Add(expCost, returnCost)
    expCost.Add(expCost, failRatio)
    return expCost

  }

  return expCost
}

func overallCosts(input formData) string {
  res := big.NewRat(0, 1)
  for i := input.Start; i < input.End; i++ {
    res.Add(res, nextStarCost(m[input.Level], i))
  }
  ac := accounting.Accounting{Symbol: "", Precision: 2}
  return ac.FormatMoney(res)
}

func handler(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()
  p := Page{""}
  if len(r.Form) > 0 {
    start, _ := strconv.Atoi(r.Form["start"][0])
    end, _ := strconv.Atoi(r.Form["end"][0])
    if end <= start {
      p.Amount = "Invalid input"
    } else {
      level, _ := strconv.Atoi(r.Form["level"][0])
      input := formData{level, start, end}
      p.Amount = "Expected cost is " + overallCosts(input)
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

