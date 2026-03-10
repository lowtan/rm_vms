<template>
  <div class="camera-container">
    <video 
      :id="'player-' + camId" 
      ref="videoPlayer" 
      autoplay 
      muted 
      playsinline
      class="video-player"
    ></video>
    
    <div v-if="isConnecting" class="overlay">
      Connecting to Camera {{ camId }}...
    </div>
    <button v-if="isConnecting&&ws" @click="reconnect">Reconnect</button>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue';
import JMuxer from 'jmuxer';

const props = defineProps({
  camId: {
    type: Number,
    required: true
  },
  wsHost: {
    type: String,
    default: 'ws://localhost:59180' // Go API Server address
  }
});

const videoPlayer = ref(null);
const isConnecting = ref(true);
let jmuxer = null;
let ws = null;

onMounted(() => {
  initPlayer();
});

onBeforeUnmount(() => {
  cleanup();
});

const reconnect = () => {
  cleanup()
  initPlayer()
}

const initPlayer = () => {
  //  Initialize JMuxer
  jmuxer = new JMuxer({
    node: 'player-' + props.camId,
    mode: 'video',          // We only have video right now
    flushingTime: 0,        // 0 = ultra-low latency. Flushes frames instantly.
    clearBuffer: true,      // Keeps memory usage low over long periods
    fps: 30,                // Fallback FPS, though the H.264 stream usually dictates this
    debug: false,
    onError: (data) => {
      console.error(`[JMuxer Cam ${props.camId}] Error:`, data);
    }
  });

  //  Initialize the WebSocket
  const wsUrl = `${props.wsHost}/ws/stream/${props.camId}`;
  ws = new WebSocket(wsUrl);

  // CRITICAL: Tell the browser we expect raw binary data, not text
  ws.binaryType = 'arraybuffer';

  ws.onopen = () => {
    console.log(`[WS] Connected to Camera ${props.camId}`);
    isConnecting.value = false;
  };

  ws.onmessage = (event) => {
    //  Feed the binary frame directly to JMuxer
    // Because your C++ worker injected the SPS/PPS, JMuxer instantly knows what to do!
    if (jmuxer) {
      jmuxer.feed({
        video: new Uint8Array(event.data)
      });
    }
  };

  ws.onclose = () => {
    console.log(`[WS] Disconnected from Camera ${props.camId}`);
    isConnecting.value = true;
    // Optional: Add reconnection logic here
  };
};

const cleanup = () => {
  if (ws) {
    ws.close();
    ws = null;
  }
  if (jmuxer) {
    jmuxer.destroy();
    jmuxer = null;
  }
};
</script>

<style>
.camera-container {
  position: relative;
  width: 100%;
  background: #000;
  aspect-ratio: 16 / 9;
}

.video-player {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.overlay {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  color: white;
  font-family: monospace;
}
</style>