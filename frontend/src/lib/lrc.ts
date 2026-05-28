export type LrcLine = {
  time: number;
  text: string;
};

const TIME_TAG_RE = /\[(\d{1,2}):(\d{2})(?:\.(\d{1,3}))?\]/g;

function parseTimeTag(minStr: string, secStr: string, fracStr?: string): number {
  const minutes = Number.parseInt(minStr, 10);
  const seconds = Number.parseInt(secStr, 10);
  let fraction = 0;
  if (fracStr) {
    const normalized = fracStr.padEnd(3, '0').slice(0, 3);
    fraction = Number.parseInt(normalized, 10) / 1000;
  }
  return minutes * 60 + seconds + fraction;
}

/** 将标准 LRC 文本解析为按时间排序的歌词行。 */
export function parseLrc(lrcText: string): LrcLine[] {
  if (!lrcText.trim()) {
    return [];
  }

  const lines: LrcLine[] = [];

  for (const rawLine of lrcText.split(/\r?\n/)) {
    const line = rawLine.trim();
    if (!line) {
      continue;
    }

    const times: number[] = [];
    TIME_TAG_RE.lastIndex = 0;
    let match: RegExpExecArray | null;
    while ((match = TIME_TAG_RE.exec(line)) !== null) {
      times.push(parseTimeTag(match[1], match[2], match[3]));
    }

    const text = line.replace(TIME_TAG_RE, '').trim();
    if (times.length === 0) {
      continue;
    }

    for (const time of times) {
      lines.push({ time, text });
    }
  }

  lines.sort((a, b) => a.time - b.time);
  return lines;
}

/** 根据当前播放时间返回应高亮的歌词行索引（-1 表示尚未开始）。 */
export function findActiveLineIndex(lines: LrcLine[], currentTime: number): number {
  if (lines.length === 0) {
    return -1;
  }

  let index = -1;
  for (let i = 0; i < lines.length; i++) {
    if (lines[i].time <= currentTime + 0.05) {
      index = i;
    } else {
      break;
    }
  }
  return index;
}
