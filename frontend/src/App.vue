<script lang="ts" setup>
import Test from './components/Test.vue'
import { NConfigProvider } from "naive-ui";
import type { GlobalThemeOverrides } from "naive-ui";

// To help with theming, uncomment this and the n-theme-editor tags below
//  import { NThemeEditor } from 'naive-ui'

const themeOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: "#008deb",
  },
  Button: {
    borderFocus: "1px solid #0082e0",
    borderHover: "1px solid #0082e0",
    textColorHover: "#0082e0",
    textColorFocus: "#0082e0",
    colorPrimary: "#008deb",
    colorHoverPrimary: "#0082e0",
    colorFocusPrimary: "#0082e0",
  },
  Input: {
    borderFocus: "1px solid #008deb",
    borderHover: "1px solid #008deb",
  },
};

</script>

<template>
  <n-config-provider :theme-overrides="themeOverrides">
    <!--    <n-theme-editor> -->
    <n-loading-bar-provider>
      <n-message-provider>
        <n-notification-provider>
          <n-dialog-provider>
            <Test />
            <n-global-style />
          </n-dialog-provider>
        </n-notification-provider>
      </n-message-provider>
    </n-loading-bar-provider>
    <!--    </n-theme-editor> -->
  </n-config-provider>

</template>

<style></style>
