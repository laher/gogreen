<script lang="ts" setup>
import { reactive } from 'vue'
import { Run, GetState, Chdir, Watch, Unwatch } from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'
import Output from './Output.vue'
import { EventsOn } from "../../wailsjs/runtime";

interface Listitem {
  label: string,
  value: string,
}

interface Data {
  list: Listitem[],
  fsnotify: boolean,
  result: string,
  cwd: string,
  disabled: boolean,
  state: main.State | null,
  testParams: main.TestParams,
}

const data: Data = reactive({
  list: [{ label: '.', value: '.' }, { label: './...', value: './...' }],
  fsnotify: false,
  result: '',
  state: null,
  cwd: '.',
  disabled: false,
  testParams: {
    pkg: './...',
    run: '',
    verbose: true,
    race: true,
  }
})

EventsOn("result", (optionalData?: any) => {
  console.log(`result received - ${optionalData}`)
  data.result = optionalData
  data.disabled = false
})

function getState() {
  GetState().then(state => {
    data.state = state
    data.cwd = state.cwd
    data.disabled = state.running || state.watching
    data.list = state.pkg_list.map(value => {
      let label = value
      // TODO make this nicer
      if (label.length > 15) {
        const parts = label.split('/')
        label = parts[parts.length - 1]
      }
      return { value, label }
    })
  }).catch(err => {
    console.log('list dirs error', err)
  })
}

getState()

function chdir() {
  Chdir().then(result => {
    console.log(`changed dir to ${result}`)
    // re-fetch the list of packages
    getState()
  }).catch(err => {
    console.log('chdir error', err)
  })
}

function watch(watch: boolean) {
  console.log('watch', watch)
  if (watch) {
    start()
  } else {
    stop()
  }
}


function start() {
  data.result = ''
  if (data.state === null) {
    console.log('no state')
    return
  }

  Watch(data.testParams).then(result => {
    data.disabled = true
    console.log(result)
  }).catch(err => {
    console.log('run error', err)
  })
}

function stop() {
  Unwatch().then(result => {
    console.log(result)
    data.disabled = false
  }).catch(err => {
    console.log('stop error', err)
  })
}

function test() {
  data.result = ''
  if (data.state === null) {
    console.log('no state')
    return
  }
  Run(data.testParams).then(result => {
    console.log(result)
    data.disabled = true
  }).catch(err => {
    console.log('run error', err)
  })
}
</script>
<style scoped>
.header_ {
  background-color: #FFD230;
}

.header_PASS {
  background-color: lightgreen;
}

.header_FAIL {
  background-color: pink;
}
</style>
<template>
  <main>
    <n-layout style="height: 768px">
      <n-layout-header style="height: 64px; padding: 24px;" bordered :class="`header_${data.result}`">
        <n-row gutter="12">
          <n-col :span="12">
            <n-input-group>
              <n-input-group-label><b>gogreen</b> </n-input-group-label>
              <n-button type="primary" @click="test" :disabled="data.disabled">Test</n-button>
            </n-input-group>
          </n-col>
          <n-col :span="12">
            <n-input-group>
              <n-switch v-model:value="data.fsnotify" @update:value="watch">
                <template #checked>
                  Watching
                </template>
                <template #unchecked>
                  Not watching
                </template>
              </n-switch>
              <!-- <n-button type="primary" @click="start">Start</n-button> -->
              <!-- <n-button type="error" @click="stop">Stop</n-button> -->
            </n-input-group>
          </n-col>
        </n-row>

      </n-layout-header>


      <n-layout position="absolute" style="top: 64px; bottom: 64px" has-sider>
        <n-layout-sider content-style="padding: 24px;" :native-scrollbar="false" collapse-mode="transform"
          :collapsed-width="20" :width="340" show-trigger="arrow-circle" bordered>
          <n-card>
            <n-h3><code>go test</code> options</n-h3>
            <n-input-group>
              <n-input-group-label>Pkg</n-input-group-label>
              <n-select v-model:value="data.testParams.pkg" :options="data.list" :disabled="data.disabled" />
            </n-input-group>
            <n-input-group>
              <n-input-group-label>Root</n-input-group-label>
              <n-input readonly v-model:value="data.cwd" :disabled="data.disabled" />
              <n-button primary @click="chdir" :disabled="data.disabled">Chdir</n-button>
            </n-input-group>
            <n-input-group>
              <n-input-group-label>Which tests</n-input-group-label>
              <n-input v-model:value="data.testParams.run" :disabled="data.disabled" />
            </n-input-group>
            <n-input-group>
              <n-checkbox v-model:checked="data.testParams.verbose" :disabled="data.disabled">
                Verbose
              </n-checkbox>
              <n-checkbox v-model:checked="data.testParams.race" :disabled="data.disabled">
                Race
              </n-checkbox>
            </n-input-group>
          </n-card>

          <n-card>
            <!--
            <n-h3>File watcher options</n-h3>
            -->

          </n-card>
        </n-layout-sider>
        <n-layout content-style="padding: 24px; min-height:600px" :native-scrollbar="false">
          <Output />
        </n-layout>
      </n-layout>
      <n-layout-footer position="absolute" style="height: 64px; padding: 24px" bordered>
        <i>It isn't easy staying green.</i>
      </n-layout-footer>
    </n-layout>
  </main>
</template>
