mod pile;
mod png;

use std::fs::File;
use std::io::{self, BufWriter, Write};
use std::process::ExitCode;
use std::time::Instant;

use pile::Pile;

const PALETTE: [[u8; 3]; 4] = [[18, 18, 32], [46, 98, 168], [236, 178, 62], [214, 64, 54]];
const ASCII: [char; 4] = [' ', '.', 'o', '#'];

const USAGE: &str = "usage:
  sandpile drop <grains> [options]     drop grains on one cell and let them settle
  sandpile identity <size> [options]   identity element of the size x size grid

options:
  -o, --output FILE   write a PNG instead of printing to the terminal
  -s, --scale N       pixels per cell in the PNG (default 2)
  --plain             print with ASCII characters instead of colored blocks";

struct Options {
    command: String,
    value: u64,
    output: Option<String>,
    scale: usize,
    plain: bool,
}

fn parse_args(args: impl IntoIterator<Item = String>) -> Result<Options, String> {
    let mut args = args.into_iter();
    let command = args.next().ok_or("missing command")?;
    let mut opts = Options {
        command,
        value: 0,
        output: None,
        scale: 2,
        plain: false,
    };
    let mut value = None;

    while let Some(arg) = args.next() {
        match arg.as_str() {
            "-o" | "--output" => {
                opts.output = Some(args.next().ok_or("--output needs a file name")?)
            }
            "-s" | "--scale" => {
                let s = args.next().ok_or("--scale needs a number")?;
                opts.scale = s
                    .parse()
                    .ok()
                    .filter(|&n| n > 0)
                    .ok_or(format!("invalid scale {s:?}"))?;
            }
            "--plain" => opts.plain = true,
            "-h" | "--help" => return Err(String::new()),
            _ if value.is_none() => {
                value = Some(
                    arg.replace('_', "")
                        .parse()
                        .map_err(|_| format!("not a number: {arg:?}"))?,
                )
            }
            _ => return Err(format!("unexpected argument {arg:?}")),
        }
    }
    opts.value = value.ok_or(format!("{} needs a number", opts.command))?;
    Ok(opts)
}

fn main() -> ExitCode {
    let opts = match parse_args(std::env::args().skip(1)) {
        Ok(o) => o,
        Err(msg) => {
            if !msg.is_empty() {
                eprintln!("sandpile: {msg}\n");
            }
            eprintln!("{USAGE}");
            return ExitCode::from(2);
        }
    };
    match run(&opts) {
        Ok(()) => ExitCode::SUCCESS,
        Err(e)
            if e.downcast_ref::<io::Error>()
                .is_some_and(|e| e.kind() == io::ErrorKind::BrokenPipe) =>
        {
            ExitCode::SUCCESS
        }
        Err(e) => {
            eprintln!("sandpile: {e}");
            ExitCode::FAILURE
        }
    }
}

fn run(opts: &Options) -> Result<(), Box<dyn std::error::Error>> {
    let start = Instant::now();
    let pile = match opts.command.as_str() {
        "drop" => {
            let grains = u32::try_from(opts.value).map_err(|_| "too many grains")?;
            let (pile, topples) = pile::drop_center(grains);
            eprintln!(
                "{grains} grains, {topples} topplings, {0}x{0} grid",
                pile.size()
            );
            pile
        }
        "identity" => {
            let size = opts.value as usize;
            if size == 0 || size > 4000 {
                return Err("size must be between 1 and 4000".into());
            }
            pile::identity(size)
        }
        other => return Err(format!("unknown command {other:?}").into()),
    };
    eprintln!("settled in {:.2?}", start.elapsed());

    match &opts.output {
        Some(path) => {
            save_png(&pile, opts.scale, path)?;
            eprintln!("wrote {path}");
        }
        None => print_terminal(&pile, opts.plain)?,
    }
    Ok(())
}

fn save_png(pile: &Pile, scale: usize, path: &str) -> io::Result<()> {
    let side = pile.size() * scale;
    let mut rgb = Vec::with_capacity(side * side * 3);
    for row in pile.rows() {
        let line: Vec<u8> = row
            .iter()
            .flat_map(|&h| std::iter::repeat_n(PALETTE[h as usize], scale))
            .flatten()
            .collect();
        for _ in 0..scale {
            rgb.extend_from_slice(&line);
        }
    }
    let file = BufWriter::new(File::create(path)?);
    png::write_rgb(file, side as u32, side as u32, &rgb)
}

fn print_terminal(pile: &Pile, plain: bool) -> io::Result<()> {
    let mut out = BufWriter::new(io::stdout().lock());
    for row in pile.rows() {
        for &h in row {
            if plain {
                let c = ASCII[h as usize];
                write!(out, "{c}{c}")?;
            } else {
                let [r, g, b] = PALETTE[h as usize];
                write!(out, "\x1b[48;2;{r};{g};{b}m  ")?;
            }
        }
        if !plain {
            write!(out, "\x1b[0m")?;
        }
        writeln!(out)?;
    }
    out.flush()
}
