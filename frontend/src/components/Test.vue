<script lang="ts" setup>
import { reactive } from 'vue'
import { Run, List, Chdir } from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'
import { EventsOn } from "../../wailsjs/runtime";


class result {
  Time: string;
  Action: string;
  Package: string;
  Output: string;
}

const r: result[] = [];
const data = reactive({
  pkg: './samples/pass',
  verbose: true,
  race: true,
  result: r,
  list: [{ label: './...', value: './...' }],
})


function list() {
  List().then(list => {
    data.list = list.map(p => { return { value: p, label: p } })
    data.pkg = './...'
  }).catch(err => {
    console.log('list dirs error', err)
  })
}

list()

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


function chdir() {
  Chdir().then(result => {
    console.log(`changed dir to ${result}`)
    // re-fetch the list of packages
    list()
  }).catch(err => {
    console.log('chdir error', err)
  })
}

function test() {
  const params: main.TestParams = {
    pkg: data.pkg,
    verbose: data.verbose,
    race: data.race,
  }
  Run(params).then(result => {
    data.result = result.split('\n').map(x => {
      if (x) {
        console.log('result-line', x)
        const parsed = JSON.parse(x)
        return parsed
      } else {
        return {}
      }
    })
  })
}

</script>

<template>
  <main>
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
    <n-card>
      <n-table>
        <n-tr v-for="r in data.result" :class="r.Action">
          <n-td class="date">{{ r.Time ? new Date(Date.parse(r.Time)).toLocaleTimeString('en-AU', {
            hour: 'numeric', minute: 'numeric',
            second: 'numeric', fractionalSecondDigits: 3
          }) : ''
            }}</n-td>
          <n-td v-if="!r.Output" class="msg">{{ r.Action + ' ' + r.Package }}</n-td>
          <n-td v-else class="msg">{{ r.Output }}</n-td>
        </n-tr>
      </n-table>
    </n-card>
  </main>
</template>

<style scoped>
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
