<script setup>
import TheWelcome from '../components/TheWelcome.vue'
import CameraStream from '../components/CameraStream.vue'
</script>

<template>
  <main>
    <CameraStream :camId="1" />
  </main>
</template>
