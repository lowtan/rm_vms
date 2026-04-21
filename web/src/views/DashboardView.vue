<template>
  <div class="dashboard">
    <h1>Dashboard</h1>
<!--     <div class="loader-wrap">
      <div class="loader"></div>
    </div> -->
    <button class="btn btn-primary" @click="toogleRefresh">{{ToogleRefreshTitle}}</button>
    <span>{{lastUpdate}}</span>
    <SHMMetrics :metrics="metrics"></SHMMetrics>
  </div>
</template>


<script setup>
import Logger from '@/utils/log';

const log = Logger.withPrefix("[Dashboard]");

import { ref, computed } from 'vue';

import SHMMetrics from '@/components/SHMMetrics.vue';

import API from '@/api';

const metrics = ref([]);

const UpdateData = function() {

  log.log("updating...")

  API.shmMetrics()
  .then(response=>{

    metrics.value = response.data;
    lastUpdate.value = new Date;

  })

}

let RefreshTimer = ref();

const ToogleRefreshTitle = computed(()=>{
  return RefreshTimer.value ? "Stop" : "Start"
})

const lastUpdate = ref();
const LastUpdateTime = computed(()=>{
  return lastUpdate.value
})

const toogleRefresh = function() {

  if(RefreshTimer.value) {
    clearInterval(RefreshTimer.value);
    RefreshTimer.value = undefined
  } else {
    RefreshTimer.value = setInterval(UpdateData, 2000);
  }

}

toogleRefresh()

</script>

<style>
@media (min-width: 1024px) {
  .dashboard {
    min-height: 100vh;
    /*display: flex;*/
    align-items: center;
  }
}
</style>
