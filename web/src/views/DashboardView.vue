<template>
  <div class="dashboard">
    <h1>Dashboard</h1>
    <SHMMetrics :metrics="metrics"></SHMMetrics>
  </div>
</template>


<script setup>
import Logger from '@/utils/log';

const log = Logger.withPrefix("[Dashboard]");

import { ref } from 'vue';

import SHMMetrics from '@/components/SHMMetrics.vue';

import API from '@/api';

const metrics = ref([]);

const UpdateData = function() {

  log.log("updating...")

  API.shmMetrics()
  .then(response=>{

    metrics.value = response.data;

  })

}

setInterval(UpdateData, 2000);

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
