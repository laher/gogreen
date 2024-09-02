<script lang="ts" setup>
import { reactive } from 'vue'
import { TreeOverrideNodeClickBehavior } from 'naive-ui'
import { Run, GetState, Chdir, Watch, Unwatch, GetTestFuncs } from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'
import Output from './Output.vue'
import { EventsOn } from "../../wailsjs/runtime";
import { PlayCircle } from '@vicons/fa'

interface Listitem {
  label: string,
  value: string,
}

interface TreeItem {
  label: string,
  key: string,
  children?: TreeItem[],
}


interface Data {
  list: Listitem[],
  testFuncs: TreeItem[],
  fsnotify: boolean,
  result: string,
  cwd: string,
  disabled: boolean,
  state: main.State | null,
  testParams: main.TestParams,
}

const data: Data = reactive({
  list: [{ label: '.', value: '.' }, { label: './...', value: './...' }],
  testFuncs: [],

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
  if (!data.fsnotify) {
    data.disabled = false
  }
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
    getTestFuncs()
  }).catch(err => {
    console.log('list dirs error', err)
  })
}


const override: TreeOverrideNodeClickBehavior = ({ option }) => {
  if (option.children) {
    if (option.label) {
      data.testParams.pkg = option.label
      data.testParams.run = ''
    }
    return 'toggleExpand'
  }
  if (option.label) {
    data.testParams.run = option.label
  }
  return 'default'
}

function getTestFuncs() {
  GetTestFuncs(data.testParams).then(testFuncs => {
    data.testFuncs = testFuncs.map(item => {
      console.log('oh', item)
      const children: TreeItem[] = item.testFuncs.map(child => {
        return {
          label: child.trim(),
          key: child.trim(),
        }
      })
      return {
        label: item.pkg,
        key: item.pkg,
        children,
      }
    })
  }).catch(err => {
    console.log('getTestFuncs error', err)
  })
}


function chdir() {
  Chdir().then(result => {
    console.log(`changed dir to ${result}`)
    data.testParams.pkg = '.'
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


getState()
</script>
<style scoped></style>
<template>
  <main>
    <n-layout style="height: 768px">
      <n-layout-header style="height: 64px; padding: 24px;" bordered :class="`header_${data.result}`">
        <n-row gutter="12">
          <n-col :span="3">
            <div class="header">
              <b>go</b> green
            </div>
          </n-col>
          <n-col :span="15" />
          <n-col :span="6">
            <n-input-group>
              <n-button type="info" @click="test" :disabled="data.disabled">
                <template #icon>
                  <n-icon>
                    <PlayCircle />
                  </n-icon>
                </template>
                Run</n-button>
              <div class="n-input-group-label" data-v-fcbebc900
                style="--n-bezier: cubic-bezier(0.4, 0, 0.2, 1); --n-group-label-color: rgb(250, 250, 252); --n-group-label-border: 1px solid rgb(224, 224, 230); --n-border-radius: 3px; --n-group-label-text-color: rgb(51, 54, 57); --n-font-size: 14px; --n-line-height: 1.6; --n-height: 34px;">
                Watch
                <n-switch v-model:value="data.fsnotify" @update:value="watch" :round="false">
                  <template #checked>
                    Stop
                  </template>
                  <template #unchecked>
                    Start
                  </template>
                </n-switch>
              </div>
            </n-input-group>
          </n-col>
        </n-row>

      </n-layout-header>


      <n-layout position="absolute" style="top: 64px; bottom: 64px" has-sider>
        <n-layout-sider content-style="padding: 24px;" :native-scrollbar="false" collapse-mode="transform"
          :collapsed-width="20" :width="340" show-trigger="arrow-circle" bordered>
          <n-form :disabled="data.disabled">
            <n-card>
              <n-h3><code>go test</code> options</n-h3>
              <n-input-group>
                <n-input-group-label>Root</n-input-group-label>
                <n-input readonly v-model:value="data.cwd" :disabled="data.disabled" />
                <n-button primary @click="chdir" :disabled="data.disabled">Chdir</n-button>
              </n-input-group>
              <n-input-group>
                <n-input-group-label>Pkg</n-input-group-label>
                <n-select v-model:value="data.testParams.pkg" :options="data.list" :disabled="data.disabled" />
              </n-input-group>
              <n-input-group>
                <n-checkbox v-model:checked="data.testParams.verbose" :disabled="data.disabled">
                  Verbose
                </n-checkbox>
                <n-checkbox v-model:checked="data.testParams.race" :disabled="data.disabled">
                  Race
                </n-checkbox>
              </n-input-group>
              <n-input-group>
                <n-input-group-label>Which tests</n-input-group-label>
                <n-input v-model:value="data.testParams.run" :disabled="data.disabled" />
              </n-input-group>
            </n-card>

            <n-card>
              <n-h3><code>test</code> funcs</n-h3>
              <n-tree block-line :data="data.testFuncs" expand-on-click cascade
                :override-default-node-click-behavior="override" :disabled="data.disabled" />
            </n-card>
          </n-form>
        </n-layout-sider>
        <n-layout content-style="padding: 24px; min-height:600px" :native-scrollbar="true">
          <Output />
        </n-layout>
      </n-layout>
      <n-layout-footer position="absolute" style="height: 64px; padding: 24px" bordered>
        <i>It isn't easy staying green.</i>
      </n-layout-footer>
    </n-layout>
  </main>
</template>
<style lang="css" scoped>
.header_ {
  background-color: #FFD230;
}

.header_PASS {
  background-color: rgba(27, 156, 51, 1);
}

.header_FAIL {
  background-color: rgba(170, 7, 7, 1);
}

.header {
  text-align: center;
  font-size: 18px;
  color: #fff;
  border: 2px solid rgb(27, 156, 51);
  padding: 1px;
  border-radius: 20px;
  background: rgb(170, 7, 7);
  background: linear-gradient(87deg, rgba(170, 7, 7, 1) 0%, rgba(100, 106, 29, 1) 22%, rgba(27, 156, 51, 1) 100%);
}
</style>
