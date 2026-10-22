pub const THRESHOLD: u32 = 4;

/// A square grid of sand heights. Grains toppled past the border fall off.
#[derive(Clone, Debug, PartialEq)]
pub struct Pile {
    size: usize,
    cells: Vec<u32>,
}

impl Pile {
    pub fn new(size: usize) -> Self {
        Self {
            size,
            cells: vec![0; size * size],
        }
    }

    pub fn filled(size: usize, height: u32) -> Self {
        Self {
            size,
            cells: vec![height; size * size],
        }
    }

    pub fn size(&self) -> usize {
        self.size
    }

    pub fn get(&self, x: usize, y: usize) -> u32 {
        self.cells[y * self.size + x]
    }

    pub fn rows(&self) -> impl Iterator<Item = &[u32]> {
        self.cells.chunks(self.size)
    }

    pub fn add(&mut self, x: usize, y: usize, grains: u32) {
        self.cells[y * self.size + x] += grains;
    }

    #[cfg(test)]
    pub fn grains(&self) -> u64 {
        self.cells.iter().map(|&c| c as u64).sum()
    }

    #[cfg(test)]
    pub fn is_stable(&self) -> bool {
        self.cells.iter().all(|&c| c < THRESHOLD)
    }

    pub fn touches_border(&self) -> bool {
        let n = self.size;
        (0..n).any(|i| {
            self.get(i, 0) > 0
                || self.get(i, n - 1) > 0
                || self.get(0, i) > 0
                || self.get(n - 1, i) > 0
        })
    }

    /// Topples until every cell holds fewer than four grains and returns the
    /// number of topplings. A cell holding `h` grains fires `h / 4` times at
    /// once, which reaches the same final state since the order does not matter.
    pub fn stabilize(&mut self) -> u64 {
        let n = self.size;
        let mut topples = 0u64;
        let mut fires = vec![0u32; self.cells.len()];
        loop {
            let mut any = false;
            for (f, c) in fires.iter_mut().zip(self.cells.iter_mut()) {
                *f = *c / THRESHOLD;
                *c %= THRESHOLD;
                any |= *f > 0;
            }
            if !any {
                return topples;
            }
            for y in 0..n {
                for x in 0..n {
                    let i = y * n + x;
                    let f = fires[i];
                    if f == 0 {
                        continue;
                    }
                    topples += f as u64;
                    if x > 0 {
                        self.cells[i - 1] += f;
                    }
                    if x + 1 < n {
                        self.cells[i + 1] += f;
                    }
                    if y > 0 {
                        self.cells[i - n] += f;
                    }
                    if y + 1 < n {
                        self.cells[i + n] += f;
                    }
                }
            }
        }
    }

    /// Adds two piles cell by cell and stabilizes the result.
    #[cfg(test)]
    pub fn sum(&self, other: &Pile) -> Pile {
        assert_eq!(self.size, other.size);
        let mut out = Pile {
            size: self.size,
            cells: self
                .cells
                .iter()
                .zip(&other.cells)
                .map(|(a, b)| a + b)
                .collect(),
        };
        out.stabilize();
        out
    }
}

/// Drops `grains` on the center of a grid just large enough to hold the
/// stabilized pile without losing any sand.
pub fn drop_center(grains: u32) -> (Pile, u64) {
    // A settled pile averages a little over two grains per cell.
    let radius = ((grains as f64) / (2.0 * std::f64::consts::PI)).sqrt();
    let mut size = (radius * 2.0) as usize + 5;
    loop {
        size |= 1;
        let mut pile = Pile::new(size);
        pile.add(size / 2, size / 2, grains);
        let topples = pile.stabilize();
        if !pile.touches_border() {
            return (pile, topples);
        }
        size = size * 5 / 4 + 4;
    }
}

/// The identity of the sandpile group on an n x n grid: the unique recurrent
/// configuration e with e + e = e. Computed as stab(6 - stab(6)).
pub fn identity(size: usize) -> Pile {
    let mut six = Pile::filled(size, 6);
    six.stabilize();
    let mut e = Pile {
        size,
        cells: six.cells.iter().map(|&c| 6 - c).collect(),
    };
    e.stabilize();
    e
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn single_topple() {
        let mut p = Pile::new(3);
        p.add(1, 1, 4);
        assert_eq!(p.stabilize(), 1);
        assert_eq!(p.get(1, 1), 0);
        assert_eq!(p.get(0, 1), 1);
        assert_eq!(p.get(1, 0), 1);
        assert_eq!(p.grains(), 4);
    }

    #[test]
    fn grains_fall_off_edges() {
        let mut p = Pile::new(1);
        p.add(0, 0, 9);
        p.stabilize();
        assert_eq!(p.get(0, 0), 1);
    }

    #[test]
    fn drop_conserves_sand() {
        for n in [1, 4, 17, 1000, 30_000] {
            let (p, _) = drop_center(n);
            assert!(p.is_stable());
            assert_eq!(p.grains(), n as u64);
        }
    }

    #[test]
    fn drop_is_symmetric() {
        let (p, _) = drop_center(5000);
        let n = p.size();
        for y in 0..n {
            for x in 0..n {
                assert_eq!(p.get(x, y), p.get(y, x));
                assert_eq!(p.get(x, y), p.get(n - 1 - x, y));
            }
        }
    }

    #[test]
    fn identity_is_idempotent() {
        for n in [1, 2, 5, 16] {
            let e = identity(n);
            assert!(e.is_stable());
            assert_eq!(e.sum(&e), e);
        }
    }

    #[test]
    fn identity_of_small_grid() {
        let e = identity(2);
        assert_eq!(e.cells, vec![2, 2, 2, 2]);
    }
}
