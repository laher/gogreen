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
  scroll()
})

EventsOn("stderr", (optionalData?: any) => {
  console.log(`data received on stderr - ${optionalData}`)
  const parsed = JSON.parse(optionalData)
  data.result.push(parsed)
  scroll()
})

const scroll = () => {
  const below = document.getElementById('below-timeline');
  if (!below) return
  setTimeout(() => {
    below.scrollIntoView({ behavior: 'smooth' });
  }, 100);

}

EventsOn("cls", () => {
  data.result = []
})

const fmtTime = (t: string) => {
  if (t === '') return ''
  return new Date(Date.parse(t)).toLocaleTimeString('en-AU', {
    hour: 'numeric', minute: 'numeric',
    second: 'numeric', fractionalSecondDigits: 3
  })
}


</script>

<template>
  <n-card id="output-card" title="Output">
    <n-timeline item-placement="left" size="large">
      <n-timeline-item v-for="r in data.result"
        :type="r.Action === 'fail' ? 'error' : r.Action === 'pass' ? 'success' : r.Action === 'start' ? 'info' : ''"
        :time="fmtTime(r.Time)" :title="r.Action === 'output' ? '' : r.Action"
        :content="r.Output ? r.Output : 'pkg ' + r.Package" />


    </n-timeline>

    <!--
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
    -->
  </n-card>
  <n-card id="below-timeline" />
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
