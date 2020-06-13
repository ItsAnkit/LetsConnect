import Vue from 'vue'
import Router from './router.js'
import App from './App.vue'
import VueRouter from 'vue-router'
import store from './store/index.js'

import BootstrapVue from 'bootstrap-vue'
import "bootstrap/dist/css/bootstrap.min.css"
import "bootstrap-vue/dist/bootstrap-vue.css"

Vue.config.productionTip = false
Vue.use(BootstrapVue)
Vue.use(VueRouter)

// new Vue({
//   render: h => h(Router),
// })

new Vue({
  // el: '#app',
  router: Router,
  store,
  template: '<App/>',
  components: { App }
}).$mount('#app')
