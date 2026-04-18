
const DefaultOpts = {
  locale: 'en',
  zoomMin: 1000 * 60,           // Zoom in up to 1 minute
  zoomMax: 1000 * 60 * 60 * 24, // Zoom out up to 24 hours
  showCurrentTime: false,       // Turn off the default red line
  format: {
    minorLabels: { minute: 'HH:mm', hour: 'HH:mm' }
  }
}

// const locales = {
//     'zh-tw': {
//       current: '目前',
//       time: '時間',
//       delete: '刪除'
//     }
// };

// // Override the axis labels to use pure numbers (e.g., 04-18 instead of Apr 18)
// // This bypasses the need for moment.js language packs completely.
// const tlformats = {
//   minorLabels: { 
//     minute: 'HH:mm', 
//     hour: 'HH:mm',
//     weekday: 'MM-DD', // Replaces English day abbreviations
//     day: 'MM-DD',     // Replaces English day abbreviations
//     month: 'YYYY-MM',
//     year: 'YYYY'
//   },
//   majorLabels: {
//     millisecond: 'HH:mm:ss',
//     second: 'MM-DD HH:mm',
//     minute: 'MM-DD',
//     hour: 'MM-DD',
//     weekday: 'YYYY-MM',
//     day: 'YYYY-MM',
//     month: 'YYYY',
//     year: ''
//   }
// }

export default DefaultOpts