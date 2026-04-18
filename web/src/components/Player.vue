<template>
  <div class="video-player-view">
    <h2>NVR Playback</h2>
    <input type="date" :value="selectDayStr" @change="onSelectDayChange">
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

import { ref, computed } from 'vue';
import Timeline from './Timeline.vue';

import TimelineDefaults from '@/data/timeline'

import API from '@/api';
import {
  WebTime,
  APITime,
  APIDayRange,
  WebTimelineBoundaries,
  ToDateStr,
} from '@/utils/time'


let selectDay = ref(new Date);
const selectDayStr = computed(() => {
  return ToDateStr(selectDay.value);
});


const fetchTimeline = function(day) {

  // let ll = log.lin("[fetchTimeline]");

  // ll.log(day);

  day = APIDayRange(day);

  API.timeline(1, day.start, day.end)
  .then(response=>{

    let data = response.data ?? {}
    let list = data.timelines ?? []

    updateTimelineItems(list);

  })

}

fetchTimeline(selectDay.value);

const Timeline2Items = function(apitl) {

  let start = APITime(apitl.start_time).WebTime().Timeline();
  let end = APITime(apitl.end_time).WebTime().Timeline();

  return {
    id: apitl.start_time,
    content: '',
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
const timelineItems = ref([]);

// Configure the view for a 24-hour period
const timelineOptionsByDate = function(date) {
  let bounaries = WebTimelineBoundaries(date);
  return {
    ...TimelineDefaults,
    ...bounaries,
  }
}

const updateTimelineBounds = function(date) {

  let bounaries = WebTimelineBoundaries(date);

  // Shift the timeline's visual window to the newly selected day
  timelineOptions.value = {
    ...timelineOptions.value,
    ...bounaries,
  };

};

const timelineOptions = ref(timelineOptionsByDate(selectDay.value));


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

  const newDate = new Date(`${e.target.value}T00:00:00`);
  selectDay.value = newDate;

  ll.log("new date:", newDate);

  updateTimelineBounds(newDate);
  fetchTimeline(newDate);

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