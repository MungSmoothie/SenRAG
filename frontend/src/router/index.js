import { createRouter, createWebHistory } from 'vue-router'
import Home from '../views/Home.vue'
import Config from '../views/Config.vue'

const routes = [
  { path: '/', name: 'Home', component: Home },
  { path: '/config', name: 'Config', component: Config }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
