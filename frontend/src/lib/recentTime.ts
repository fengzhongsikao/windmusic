/** 将 ISO 时间格式化为相对时间（中文） */
export function formatPlayedAt(iso: string): string {
  const played = new Date(iso);
  if (Number.isNaN(played.getTime())) {
    return '—';
  }
  const diffMs = Date.now() - played.getTime();
  if (diffMs < 0) {
    return '刚刚';
  }
  const minutes = Math.floor(diffMs / 60_000);
  if (minutes < 1) {
    return '刚刚';
  }
  if (minutes < 60) {
    return `${minutes} 分钟前`;
  }
  const hours = Math.floor(minutes / 60);
  if (hours < 24) {
    return `${hours} 小时前`;
  }
  const days = Math.floor(hours / 24);
  if (days === 1) {
    return '昨天';
  }
  if (days < 7) {
    return `${days} 天前`;
  }
  return played.toLocaleDateString('zh-CN', { month: 'numeric', day: 'numeric' });
}
