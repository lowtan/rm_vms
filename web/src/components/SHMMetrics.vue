<template>
  <div class="workers-grid">
    <div v-for="worker in metrics" :key="worker.worker_id" class="worker-card">
      <h3 class="worker-title">
        <IconServer/>
        Worker {{ worker.worker_id }}
      </h3>

      <div class="channels-list">
        <div v-for="(channel, idx) in worker.channels" :key="idx" class="channel-row">
          
          <div class="channel-info">
            <span class="cam-badge">Cam {{ channel.cam_id }}</span>
            <span class="bytes-text">{{ formatBytes(channel.bytes_buffered) }} / {{ formatBytes(channel.capacity) }}</span>
          </div>

          <div class="progress-container">
            <div 
              class="progress-bar" 
              :class="getSaturationColor(channel.saturation_pct, channel.is_stalled)"
              :style="{ width: Math.min(channel.saturation_pct, 100) + '%' }"
            ></div>
          </div>

          <div class="channel-stats">
            <span class="pct-text">{{ channel.saturation_pct.toFixed(2) }}%</span>
            <span v-if="channel.is_stalled" class="stalled-warning">STALLED</span>
          </div>

        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { defineProps } from 'vue'

import IconServer from './icons/IconServer.vue'

// Pure presentational component: only reads data passed from the parent
const props = defineProps({
  metrics: {
    type: Array,
    required: true,
    default: () => []
  }
})

// Color coding based on buffer pressure
const getSaturationColor = (pct, isStalled) => {
  if (isStalled) return 'bg-stalled'
  if (pct > 80) return 'bg-danger'
  if (pct > 50) return 'bg-warning'
  return 'bg-success'
}

// Human-readable byte formatting
const formatBytes = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}
</script>

<style scoped>
.workers-grid {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.worker-card {
  background-color: #181825;
  border-radius: 8px;
  padding: 16px;
  border: 1px solid #313244;
  color: #cdd6f4;
  font-family: system-ui, -apple-system, sans-serif;
}

.worker-title {
  margin: 0 0 16px 0;
  font-size: 1.2rem;
  color: #89b4fa;
  display: flex;
  align-items: center;
  gap: 8px;
}

.channels-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.channel-row {
  display: grid;
  grid-template-columns: 180px 1fr 100px;
  align-items: center;
  gap: 16px;
  background-color: #1e1e2e;
  padding: 10px;
  border-radius: 6px;
}

.channel-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.cam-badge {
  font-weight: bold;
  font-size: 0.9rem;
  color: #f5e0dc;
}

.bytes-text {
  font-size: 0.75rem;
  color: #a6adc8;
}

.progress-container {
  height: 8px;
  background-color: #313244;
  border-radius: 4px;
  overflow: hidden;
}

.progress-bar {
  height: 100%;
  transition: width 0.3s ease, background-color 0.3s ease;
}

.bg-success { background-color: #a6e3a1; }
.bg-warning { background-color: #f9e2af; }
.bg-danger { background-color: #fab387; }
.bg-stalled { background-color: #f38ba8; animation: pulse 1s infinite; }

.channel-stats {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
}

.pct-text {
  font-family: monospace;
  font-size: 0.9rem;
}

.stalled-warning {
  font-size: 0.7rem;
  font-weight: bold;
  color: #181825;
  background-color: #f38ba8;
  padding: 2px 6px;
  border-radius: 4px;
}

@keyframes pulse {
  0% { opacity: 1; }
  50% { opacity: 0.6; }
  100% { opacity: 1; }
}
</style>