<template>
  <div class="loader-wrap">
    <div class="loader"></div>
    <aside v-if="message">{{message}}</aside>
  </div>
</template>

<script setup>
import { defineProps } from 'vue'

const props = defineProps({
  message: {
    type: String,
    required: false,
  }
})


</script>