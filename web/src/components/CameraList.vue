<template>
  <div class="row g-4">
    
    <div v-if="cameras.length === 0" class="col-12 text-center text-muted py-5">
      No cameras configured or loading...
    </div>

    <div v-for="cam in cameras" :key="cam.id" class="col-12 col-md-6 col-xl-4">
        <CameraCard :camera="cam"></CameraCard>
    </div>

  </div>
</template>

<script setup>
import { defineProps } from 'vue'
import CameraCard from './CameraCard.vue'

// Pure presentational component
const props = defineProps({
  cameras: {
    type: Array,
    required: true,
    default: () => []
  }
})
</script>
