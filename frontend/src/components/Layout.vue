<script lang="ts" setup>
import { reactive } from 'vue'
import { Run, List, Chdir, StopFSNotify, StartFSNotify } from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'
import Test from './Test.vue'

const data = reactive({
  pkg: './samples/pass',
  verbose: true,
  race: true,
  list: [{ label: './...', value: './...' }],
  fsnotify: false,
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

function chdir() {
  Chdir().then(result => {
    console.log(`changed dir to ${result}`)
    // re-fetch the list of packages
    list()
  }).catch(err => {
    console.log('chdir error', err)
  })
}


function start() {
  const params: main.TestParams = {
    pkg: data.pkg,
    verbose: data.verbose,
    race: data.race,
  }
  StartFSNotify(params).then(result => {
    console.log(result)
  }).catch(err => {
    console.log('run error', err)
  })
}

function stop() {
  StopFSNotify().then(result => {
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
  }
  Run(params).then(result => {
    console.log(result)
  }).catch(err => {
    console.log('run error', err)
  })
}
</script>
<template>
  <n-main>
    <n-layout style="height: 768px">
      <n-layout-header style="height: 64px; padding: 24px; background-color: #FFD230" bordered>
        <n-row gutter="12">
          <n-col :span="12">
            <n-input-group>
              <n-input-group-label><b>gotesty</b> </n-input-group-label>
              <n-button type="primary" @click="test">Run</n-button>
            </n-input-group>
          </n-col>
          <n-col :span="12">
            <n-input-group>
              <n-input-group-label>watch</n-input-group-label>
              <n-switch v-model:value="data.fsnotify" />
              <!-- <n-button type="primary" @click="start">Start</n-button> -->
              <!-- <n-button type="error" @click="stop">Stop</n-button> -->
            </n-input-group>
          </n-col>
        </n-row>
      </n-layout-header>


      <n-layout position="absolute" style="top: 64px; bottom: 64px" has-sider>
        <n-layout-sider content-style="padding: 24px;" :native-scrollbar="false" collapse-mode="transform"
          :collapsed-width="120" :width="240" show-trigger="arrow-circle" bordered>
          <n-h2>Options</n-h2>
          <n-card>
            <n-input-group>

              <n-dropdown trigger="hover" v-model:value="data.pkg" :options="data.list">
                <n-button>Package</n-button>
              </n-dropdown>
            </n-input-group>
            <n-input-group>
              <n-button primary @click="chdir">Chdir</n-button>
            </n-input-group>
            <n-input-group>
              <n-checkbox v-model:checked="data.verbose">
                Verbose
              </n-checkbox>
            </n-input-group>
            <n-input-group>
              <n-checkbox v-model:checked="data.race">
                Race
              </n-checkbox>
            </n-input-group>
          </n-card>
        </n-layout-sider>
        <n-layout content-style="padding: 24px; min-height:600px" :native-scrollbar="false">
          <Test />
        </n-layout>
      </n-layout>
      <n-layout-footer position="absolute" style="height: 64px; padding: 24px" bordered>
        <i>Run your tests like the moon pulls the tides - always</i>
      </n-layout-footer>
    </n-layout>
  </n-main>
</template>
