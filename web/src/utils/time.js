import { format, getUnixTime, fromUnixTime, startOfDay, endOfDay } from 'date-fns';

/**
 * ========================================================
 * Date Time Format Converting
 * ========================================================
 */

// getUnixTime automatically converts milliseconds to seconds
const DatetimeToAPIStamps = function(date) {
    return getUnixTime(date);
}

// fromUnixTime automatically converts seconds back to a Date object
const APIStampsToDatetime = function(stamps) {
    return fromUnixTime(stamps);
}

// Safely formats to '2026-04-15T08:30:00', handling all zero-padding
const ToTimelineStr = function(date) {
    return format(date, "yyyy-MM-dd'T'HH:mm:ss");
}

const TimeCVT = {
    DatetimeToAPIStamps,
    APIStampsToDatetime,
    ToTimelineStr,
}

/**
 * Object-Oriented Wrapper
 * Allows chaining or contextual operations around a specific Date.
 * Added a default parameter so Time() safely defaults to now.
 */
const Time = function(date = new Date()) {
    return {
        APIStamp: () => getUnixTime(date),
        Timeline: () => format(date, "yyyy-MM-dd'T'HH:mm:ss"),
        Native: () => date // Always good to have an escape hatch to get the raw Date back
    }
}


/**
 * ========================================================
 * Day Ranging
 * ========================================================
 */
const DayRange = function(date) {
    // startOfDay and endOfDay safely handle exact boundary calculations
    let start = getUnixTime(startOfDay(date));
    let end = getUnixTime(endOfDay(date));

    return { start, end }
}

const TodayRange = function() {
    return DayRange(new Date)
}

export {
    Time,
    TimeCVT,
    DayRange,
    TodayRange
}