# Starforcing Calculator

[Try it here!](https://starforce.appspot.com/)

## Background

Formulas taken from [here](https://amph.shinyapps.io/starforce/_w_0d9691a6/starforce.pdf) (note that there is a typo in the formulas for 15*->16* and 20*->21* and they should have the s<sub>15</sub> and s<sub>20</sub> beneath all terms) Credit to reddit user u/hailcrest for these derivations.

Starforcing is the enhancement system for equipment in MapleStory. If you are unfamiliar with MapleStory you can think of starforcing as a
statistical game like this:

We are playing game with 25 levels. In MapleStory, the first ten levels are mostly trivial, so for convenience's sake, we'll have you start at level 10.
The objective is to increase your levels as much as possible. At each level ![equation](https://latex.codecogs.com/gif.latex?i), we can attempt to go up to level
![equation](https://latex.codecogs.com/gif.latex?i&plus;1) for a cost ![equation](https://latex.codecogs.com/gif.latex?c_i) with a probability of success ![equation](https://latex.codecogs.com/gif.latex?s_i)
If you fail, with a chance ![equation](https://latex.codecogs.com/gif.latex?f_i) (note this is not necessarily ![equation](https://latex.codecogs.com/gif.latex?1-s_i)), you go down a level to level ![equation](https://latex.codecogs.com/gif.latex?f_i), except when at levels 10, 15, and 20, in in which case you stay at level 10, 15, or 20 respectively. In addition, if you drop two levels in a row, your success rate for the next attempt becomes 100%.

Starting from the level 12->13 enhancement, you also have the chance to lose your progress, and reset back to level 12. We denote this probability as ![equation](https://latex.codecogs.com/gif.latex?d_i). For example, when attempting to go from level 15->16 or level 21->22, we have a ![equation](https://latex.codecogs.com/gif.latex?d_%7B15%7D) and ![equation](https://latex.codecogs.com/gif.latex?d_%7B21%7D) chance of instantly dropping back down to level 12. However, starting from the level 12->13 enhancement and up to the level 16->17 enhancement, you can also choose to
pay ![equation](https://latex.codecogs.com/gif.latex?2c_i) in order to remove this chance. Instead, your success rate will still be ![equation](https://latex.codecogs.com/gif.latex?s_i) and your failure rate will be ![equation](https://latex.codecogs.com/gif.latex?1-s_i).

The goal of this calculator is to provide an expected cost of going from one level to another.

## Use

Select all the paremeters that apply to you, then press "Go!" when you are ready to see the results!


