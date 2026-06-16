// dowLabel returns the short weekday name for day-of-week index d (0=Sun…6=Sat)
// using Jan 1 2023 as the anchor date (a Sunday). The formatter must pin
// timeZone:'UTC' so the anchor date doesn't shift back a day on hosts west of UTC.
export const dowLabel = (d: number, lang: string): string =>
  new Intl.DateTimeFormat(lang, { weekday: 'short', timeZone: 'UTC' }).format(new Date(Date.UTC(2023, 0, 1 + d)))
