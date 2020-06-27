import Vue from 'vue'
import Vuex from "vuex"
import createPersistedState from 'vuex-persistedstate'

Vue.use(Vuex)

export default new Vuex.Store({
    plugins: [createPersistedState({
        storage: window.sessionStorage
    })],
    state: { 
        user: {
            id: '',
            username: '',
            mobile_no: '',
        }
    },
    getters: {
        getUser: state => {
            return state.user
        },
    },
    mutations: {
        updateUser (state, value) {
            state.user = value
        },
        clearStore (state) {
            state.user = {}
        },
    },
    actions: {}
})