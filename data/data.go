package data

import (
  "io/ioutil"
  "gopkg.in/yaml.v2"
  "bytes"
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


