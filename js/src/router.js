import Vue from 'vue'
import Router from 'vue-router'
import Chatroom from './components/Chatroom.vue'
import Login from './components/Login.vue'

Vue.use(Router)

export default new Router({
  mode: 'history',
  routes: [
    {
      path: '/',
      name: 'Login',
      component: Login
    },
    {
      path: '/chatroom',
      name: 'Chatroom',
      component: Chatroom
    }
  ]
})
