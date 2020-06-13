import Vue from 'vue'
import Vuex from "vuex"

Vue.use(Vuex)

export default new Vuex.Store({
    state: { 
        username: "Professor"
    },
    getters: {
        getUsername: state => {
            return state.username
        }
    },
    mutations: {
        updateUsername (state, value) {
            state.username = value
        }
    },
    actions: {}
})