<script lang="ts" setup>
import { reactive } from 'vue'
import { EventsOn } from "../../wailsjs/runtime";


class result {
  Time: string;
  Action: string;
  Package: string;
  Output: string;
}

const r: result[] = [];
const data = reactive({
  result: r,
})


EventsOn("stdout", (optionalData?: any) => {
  console.log(`data received on stdout - ${optionalData}`)
  const parsed = JSON.parse(optionalData)
  data.result.push(parsed)
})

EventsOn("stderr", (optionalData?: any) => {
  console.log(`data received on stderr - ${optionalData}`)
  const parsed = JSON.parse(optionalData)
  data.result.push(parsed)
})

EventsOn("cls", () => {
  data.result = []
})


</script>

<template>
  <!--
  <n-card>
    <n-input-group>
      <n-button primary @click="chdir">Chdir</n-button>
      <n-input-group-label>Package</n-input-group-label>
      <n-select v-model:value="data.pkg" :options="data.list" :style="{ width: '33%' }" />
      <n-button primary @click="test">Test</n-button>
    </n-input-group>
    <n-input-group>
      <n-checkbox v-model:checked="data.verbose">
        Verbose
      </n-checkbox>
      <n-checkbox v-model:checked="data.race">
        Race
      </n-checkbox>
    </n-input-group>
  </n-card>
    -->
  <n-card>
    <n-table>
      <n-tr v-for="r in data.result" :class="r.Action">
        <n-th class="date">{{ r.Time ? new Date(Date.parse(r.Time)).toLocaleTimeString('en-AU', {
          hour: 'numeric', minute: 'numeric',
          second: 'numeric', fractionalSecondDigits: 3
        }) : ''
          }}</n-th>
        <n-td v-if="!r.Output" class="msg">{{ r.Action + ' ' + r.Package }}</n-td>
        <n-td v-else class="msg">{{ r.Output }}</n-td>
      </n-tr>
    </n-table>
  </n-card>
</template>

<style scoped>
th {
  max-width: '10%';
}

.run td {
  background-color: cyan;
}

.output td.date {
  background-color: cyan;
}

.start td {
  background-color: cyan;
}

.test td {
  background-color: lightgrey;
}


.fail td {
  background-color: lightpink;
}

.pass td {
  background-color: lightgreen;
}
</style>
