use std::io::{self, Write};

/// Writes an 8-bit RGB PNG. Pixel data goes into uncompressed deflate blocks,
/// which keeps the encoder tiny at the cost of file size.
pub fn write_rgb<W: Write>(mut w: W, width: u32, height: u32, rgb: &[u8]) -> io::Result<()> {
    assert_eq!(rgb.len(), (width * height * 3) as usize);

    w.write_all(b"\x89PNG\r\n\x1a\n")?;

    let mut ihdr = Vec::with_capacity(13);
    ihdr.extend_from_slice(&width.to_be_bytes());
    ihdr.extend_from_slice(&height.to_be_bytes());
    ihdr.extend_from_slice(&[8, 2, 0, 0, 0]);
    chunk(&mut w, b"IHDR", &ihdr)?;

    let stride = width as usize * 3;
    let mut raw = Vec::with_capacity((stride + 1) * height as usize);
    for row in rgb.chunks(stride) {
        raw.push(0);
        raw.extend_from_slice(row);
    }
    chunk(&mut w, b"IDAT", &zlib_stored(&raw))?;
    chunk(&mut w, b"IEND", &[])
}

fn chunk<W: Write>(w: &mut W, kind: &[u8; 4], data: &[u8]) -> io::Result<()> {
    w.write_all(&(data.len() as u32).to_be_bytes())?;
    w.write_all(kind)?;
    w.write_all(data)?;
    let crc = crc32(kind.iter().chain(data));
    w.write_all(&crc.to_be_bytes())
}

fn zlib_stored(data: &[u8]) -> Vec<u8> {
    const MAX_BLOCK: usize = 65_535;
    let mut out = Vec::with_capacity(data.len() + data.len() / MAX_BLOCK * 5 + 11);
    out.extend_from_slice(&[0x78, 0x01]);

    let blocks: Vec<&[u8]> = if data.is_empty() {
        vec![&[]]
    } else {
        data.chunks(MAX_BLOCK).collect()
    };
    for (i, block) in blocks.iter().enumerate() {
        out.push((i + 1 == blocks.len()) as u8);
        let len = block.len() as u16;
        out.extend_from_slice(&len.to_le_bytes());
        out.extend_from_slice(&(!len).to_le_bytes());
        out.extend_from_slice(block);
    }
    out.extend_from_slice(&adler32(data).to_be_bytes());
    out
}

fn crc32<'a>(bytes: impl IntoIterator<Item = &'a u8>) -> u32 {
    let mut crc = 0xFFFF_FFFFu32;
    for &b in bytes {
        crc ^= b as u32;
        for _ in 0..8 {
            crc = if crc & 1 != 0 {
                (crc >> 1) ^ 0xEDB8_8320
            } else {
                crc >> 1
            };
        }
    }
    !crc
}

fn adler32(data: &[u8]) -> u32 {
    let (mut a, mut b) = (1u32, 0u32);
    for chunk in data.chunks(5552) {
        for &x in chunk {
            a += x as u32;
            b += a;
        }
        a %= 65_521;
        b %= 65_521;
    }
    (b << 16) | a
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn checksums_match_reference_values() {
        assert_eq!(crc32(b"123456789"), 0xCBF4_3926);
        assert_eq!(adler32(b"Wikipedia"), 0x11E6_0398);
    }

    #[test]
    fn writes_valid_header() {
        let mut buf = Vec::new();
        write_rgb(&mut buf, 2, 1, &[255, 0, 0, 0, 0, 255]).unwrap();
        assert_eq!(&buf[..8], b"\x89PNG\r\n\x1a\n");
        assert_eq!(&buf[12..16], b"IHDR");
        assert_eq!(&buf[buf.len() - 8..buf.len() - 4], b"IEND");
    }
}
