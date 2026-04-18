<template>
  <div class="video-player-view">
    <h2>NVR Playback</h2>
    <input type="date" :value="selectDay" @change="onSelectDayChange">
    <video src="" class="player-video"></video>
    <Timeline 
      :initialTime="new Date"
      :items="timelineItems" 
      :options="timelineOptions"
      @timechange="handleScrubbing"
      @seek="handleUserSeek"
    />
  </div>
</template>

<script setup>

import Logger from '@/utils/log';

const log = Logger.withPrefix("[Player]");

import { ref } from 'vue';
import Timeline from './Timeline.vue';

import API from '@/api';
import {
  TimeCVT,
  DayRange,
  TodayRange,
} from '@/utils/time'

let selectDay = ref(new Date);

// let today = TodayRange();

const fetchTimeline = function(day) {

  let ll = log.lin("[fetchTimeline]");

  ll.log(day);

  day = DayRange(day);

  API.timeline(1, day.start, day.end)
  .then(response=>{

    let data = response.data ?? {}
    let list = data.timelines ?? []
    ll.log("timeline:", list.length);

    updateTimelineItems(list);

  })

}

fetchTimeline(selectDay.value);

const Timeline2Items = function(tl) {

  let ll = log.lin("[Timeline2Items]");

  let start = TimeCVT.ToTimelineStr(TimeCVT.APIStampsToDatetime(tl.start_time))
  let end = TimeCVT.ToTimelineStr(TimeCVT.APIStampsToDatetime(tl.end_time))

  ll.log("time:", start, end)

  return {
    id: tl.start_time,
    content: 'Continuous',
    start: start,
    end: end,
    type: 'background',
    className: 'continuous-record'
  }
}

const updateTimelineItems = function(list) {

  let ll = log.lin("[updateTimelineItems]");
  ll.log("new list length", list.length)

  while(timelineItems.value.length) {
    timelineItems.value.shift(); 
  }
  list.map((v)=>{
    timelineItems.value.push(Timeline2Items(v))
  })

}


// Define motion events or continuous recording blocks
const timelineItems = ref([
  { id: 1, content: 'Continuous', start: '2026-04-15T00:00:00', end: '2026-04-15T08:30:00', type: 'background', className: 'continuous-record' },
  { id: 2, content: 'Motion', start: '2026-04-15T14:15:00', end: '2026-04-15T14:17:30', style: 'background-color: red;' }
]);

// Configure the view for a 24-hour period
const timelineOptions = ref({
  start: '2026-04-18T00:00:00', // Today 00:00:00
  end: '2026-04-18T23:59:59',   // Today 23:59:59
  min: '2026-04-17T23:00:00',   // Prevent panning too far back
  max: '2026-04-19T01:00:00',   // Prevent panning too far forward
  zoomMin: 1000 * 60,           // Zoom in up to 1 minute
  zoomMax: 1000 * 60 * 60 * 24, // Zoom out up to 24 hours
  showCurrentTime: false,       // Turn off the default red line
  format: {
    minorLabels: { minute: 'h:mm', hour: 'hh' }
  }
});

// Handle the user dragging the playhead
const handleScrubbing = (properties) => {

  let ll = log.lin("[handleScrubbing]");

  const scrubbedTime = properties.time;
  // ll.log("User is scrubbing to:", scrubbedTime);

  // Send this timestamp over WebSocket to your Go backend to fetch the new video chunk


};

const handleUserSeek = (seek) => {

  let ll = log.lin("[handleUserSeek]");
  ll.log(seek);

}


const onSelectDayChange = (e)=> {

  let ll = log.lin("[onSelectDayChange]");

  ll.log(e.target);

  fetchTimeline(new Date(e.target.value));

}

</script>

<style>
  video.player-video {
    width: 100%;
    min-height: 300px;
    min-width: 300px;
    border: 1px solid;
  }
</style>