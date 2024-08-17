import { createApp } from 'vue'
import App from './App.vue'
import './style.css';
import { createPinia } from "pinia";
import naive from "naive-ui";

const pinia = createPinia();
const app = createApp(App);
app.use(naive).use(pinia);
app.mount("#app");
