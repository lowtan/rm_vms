<template>
  <div ref="timelineContainer" class="timeline-container"></div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch } from 'vue';
import { Timeline } from 'vis-timeline/standalone';
import { DataSet } from 'vis-data';

// Define the props we expect from the parent
const props = defineProps({
  items: {
    type: Array,
    required: true,
    default: () => []
  },
  groups: {
    type: Array,
    required: false,
    default: () => []
  },
  options: {
    type: Object,
    required: false,
    default: () => ({})
  },
  initialTime: {
    type: [Date, String, Number],
    required: true
  },
});

// Define the events we want to pass up to the parent Vue component
const emit = defineEmits(['select', 'timechange', 'timechanged', 'seek']);

const timelineContainer = ref(null);
let timelineInstance = null;
let itemsDataset = null;
let groupsDataset = null;

onMounted(() => {

  // Convert static arrays to vis-data DataSets for optimized rendering
  itemsDataset = new DataSet(props.items);
  if (props.groups.length > 0) {
    groupsDataset = new DataSet(props.groups);
  }

  // Initialize the timeline
  timelineInstance = new Timeline(
    timelineContainer.value,
    itemsDataset,
    groupsDataset,
    props.options
  );

  // 'playhead' is the internal ID we give this specific bar
  timelineInstance.addCustomTime(props.initialTime, 'playhead');

  // Bind vis-timeline events to Vue emits
  timelineInstance.on('select', (properties) => emit('select', properties));

  // Handle clicks on the timeline background
  timelineInstance.on('click', (properties) => {
    // properties.time is the exact Date object where the user clicked
    if (properties.time) {
      timelineInstance.setCustomTime(properties.time, 'playhead');
      emit('seek', properties.time);
    }
  });

  // 'timechange' fires repeatedly while dragging the custom time bar (scrubbing)
  timelineInstance.on('timechange', (properties) => emit('timechange', properties));

  // 'timechanged' fires once when dragging stops
  timelineInstance.on('timechanged', (properties) => emit('timechanged', properties));
});

// Watch for changes in the parent's data and update the datasets efficiently
watch(() => props.items, (newItems) => {
  if (itemsDataset) {
    itemsDataset.clear();
    itemsDataset.add(newItems);
  }
}, { deep: true });

watch(() => props.groups, (newGroups) => {
  if (groupsDataset) {
    groupsDataset.clear();
    groupsDataset.add(newGroups);
  }
}, { deep: true });

watch(() => props.options, (newOptions) => {
  if (timelineInstance) {
    timelineInstance.setOptions(newOptions);
  }
}, { deep: true });


// EXPOSE THIS TO THE PARENT
// This allows the parent component to push the playhead forward continuously 
// as the video plays, or jump it forward/backward by clicking external UI buttons.
const setPlayheadTime = (newTime) => {
  if (timelineInstance) {
    timelineInstance.setCustomTime(newTime, 'playhead');

    // Optional: Keep the timeline centered on the playhead if it moves off-screen
    // timelineInstance.moveTo(newTime, { animation: true });
  }
};

defineExpose({
  setPlayheadTime
});

// CRITICAL: Clean up the instance when the component is unmounted
onBeforeUnmount(() => {
  if (timelineInstance) {
    timelineInstance.destroy();
    timelineInstance = null;
  }
});
</script>

<style>
.timeline-container {
  width: 100%;
  /* Optional: Set a min-height so it doesn't collapse before data loads */
  min-height: 150px; 
  background: white;
}

.vis-item.vis-background.continuous-record {
  background-color: rgba(99, 190, 99, .8);
}

/* You can override default vis-timeline CSS variables here */
:deep(.vis-timeline) {
  border: 1px solid #444;
  background-color: #1e1e1e;
  color: white;
}
:deep(.vis-text) {
  color: #EEE;
}
.vis-text {
    color: #EEEEEE;
}
</style>