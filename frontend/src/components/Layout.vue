<script lang="ts" setup>
import { reactive } from 'vue'
import { Run, List, Chdir, Watch, Unwatch } from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'
import Output from './Output.vue'

const data = reactive({
  pkg: '.',
  cwd: '',
  verbose: true,
  race: true,
  list: [{ label: '.', value: '.' }, { label: './...', value: './...' }],
  fsnotify: false,
  run: '',
})

function list() {
  List().then(list => {
    data.cwd = list.cwd
    data.list = list.pkg_list.map(value => {
      let label = value
      // TODO make this nicer
      if (label.length > 15) {
        const parts = label.split('/')
        label = parts[parts.length - 1]
      }
      return { value, label }
    })
    data.pkg = '.'
  }).catch(err => {
    console.log('list dirs error', err)
  })
}

list()

function chdir() {
  Chdir().then(result => {
    console.log(`changed dir to ${result}`)
    // re-fetch the list of packages
    list()
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
  const params: main.TestParams = {
    pkg: data.pkg,
    verbose: data.verbose,
    race: data.race,
    run: data.run,
  }
  Watch(params).then(result => {
    console.log(result)
  }).catch(err => {
    console.log('run error', err)
  })
}

function stop() {
  Unwatch().then(result => {
    console.log(result)
  }).catch(err => {
    console.log('run error', err)
  })
}
function test() {
  const params: main.TestParams = {
    pkg: data.pkg,
    verbose: data.verbose,
    race: data.race,
    run: data.run,
  }
  Run(params).then(result => {
    console.log(result)
  }).catch(err => {
    console.log('run error', err)
  })
}
</script>
<template>
  <main>
    <n-layout style="height: 768px">
      <n-layout-header style="height: 64px; padding: 24px; background-color: #FFD230" bordered>
        <n-row gutter="12">
          <n-col :span="12">
            <n-input-group>
              <n-input-group-label><b>gogetgreen</b> </n-input-group-label>
              <n-button type="primary" @click="test">Test</n-button>
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
          <n-h2>Options</n-h2>
          <n-card>
            <n-h3>gotest options</n-h3>
            <n-input-group>
              <n-input-group-label>Pkg</n-input-group-label>
              <n-select v-model:value="data.pkg" :options="data.list" />
            </n-input-group>
            <n-input-group>
              <n-input-group-label>Root</n-input-group-label>
              <n-input readonly v-model:value="data.cwd" />
              <n-button primary @click="chdir">Chdir</n-button>
            </n-input-group>
            <n-input-group>
              <n-input-group-label>Which tests</n-input-group-label>
              <n-input v-model:value="data.run" />
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
            <n-h3>File watcher options</n-h3>

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
