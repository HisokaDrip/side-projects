# sandpile

A simulator for the [Abelian sandpile model](https://en.wikipedia.org/wiki/Abelian_sandpile_model),
written in Rust with no dependencies.

Every cell of a square grid holds some grains of sand. A cell with four or more
grains topples: it gives one grain to each of its four neighbours, and grains
pushed off the edge are lost. Repeat until nothing can topple. Simple rules,
surprisingly intricate fractal patterns.

## Build

```
cargo build --release
```

## Usage

Drop a pile of sand on a single cell and watch it spread:

```
$ ./target/release/sandpile drop 300 --plain
300 grains, 1685 topplings, 19x19 grid
...
            ..oo######oo..
          oo..oo..##..oo..oo
        oooo##..######..##oooo
      ....##oo####..####oo##....
      oooo..##oooo  oooo##..oooo
      ##..####oooo##oooo####..##
      ######..  ##  ##  ..######
...
```

Write a PNG instead (the grid is sized automatically so no sand is lost):

```
./target/release/sandpile drop 200_000 -o pile.png
```

Compute the identity element of the sandpile group on an n x n grid, the
configuration that leaves any stable pile unchanged when added to it:

```
./target/release/sandpile identity 300 -s 2 -o identity.png
```

| option            | meaning                                         |
|-------------------|-------------------------------------------------|
| `-o, --output F`  | write a PNG to `F` instead of printing          |
| `-s, --scale N`   | pixels per cell in the PNG (default 2)          |
| `--plain`         | ASCII output instead of 24-bit colored blocks   |

Colors from darkest to brightest mean 0, 1, 2 and 3 grains.

The number of topplings grows roughly with the square of the grain count:
100,000 grains settle in about two seconds, 1,000,000 take several minutes.

## Tests

```
cargo test
```
