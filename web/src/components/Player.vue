<template>
  <div class="video-player-view">
    <h2>NVR Playback</h2>
    <video src="" class="player-video"></video>
    <Timeline 
      :items="timelineItems" 
      :options="timelineOptions"
      @timechange="handleScrubbing"
    />
  </div>
</template>

<script setup>
import { ref } from 'vue';
import Timeline from './timeline.vue';

import API from '@/api';
import {TodayRange} from '@/utils/time'

let today = TodayRange();

console.log(today)

API.timeline(1, today.start, today.end)
.then(response=>{

  console.log("timeline:",response.timeline)

})

// Define motion events or continuous recording blocks
const timelineItems = ref([
  { id: 1, content: 'Continuous', start: '2026-04-15T00:00:00', end: '2026-04-15T08:30:00', type: 'background', className: 'continuous-record' },
  { id: 2, content: 'Motion', start: '2026-04-15T14:15:00', end: '2026-04-15T14:17:30', style: 'background-color: red;' }
]);

// Configure the view for a 24-hour period
const timelineOptions = ref({
  start: '2026-04-15T00:00:00', // Today 00:00:00
  end: '2026-04-15T23:59:59',   // Today 23:59:59
  min: '2026-04-14T23:00:00',   // Prevent panning too far back
  max: '2026-04-16T01:00:00',   // Prevent panning too far forward
  zoomMin: 1000 * 60,           // Zoom in up to 1 minute
  zoomMax: 1000 * 60 * 60 * 24, // Zoom out up to 24 hours
  showCurrentTime: false,       // Turn off the default red line
  format: {
    minorLabels: { minute: 'h:mm', hour: 'hh' }
  }
});

// Handle the user dragging the playhead
const handleScrubbing = (properties) => {

  const scrubbedTime = properties.time;
  console.log("User is scrubbing to:", scrubbedTime);
  // Send this timestamp over WebSocket to your Go backend to fetch the new video chunk

};
</script>

<style>
  video.player-video {
    width: 100%;
    min-height: 300px;
    min-width: 300px;
    border: 1px solid;
  }
</style>