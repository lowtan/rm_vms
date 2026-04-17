
const DayRange = function(day) {

    let y = day.getFullYear();
    let m = day.getMonth();
    let d = day.getDate();

    let start = new Date(y,m,d,0,0,0,0).getTime();
    let end = new Date(y,m,d,23,59,59,999).getTime();

    return {start,end}

}

const TodayRange = function() {
    return DayRange(new Date)
}

export {
    DayRange,
    TodayRange
}