<template>
    <div class="chat">
        <div class="row d-flex justify-content-center mt-4">
            <div class="input-field w-25">
                <input type="text" v-model.trim="username" placeholder="Username">
            </div>
            <div class="input-field">
                <button class="waves-effect waves-light btn" @click="join()">
                    <i class="material-icons right">done</i>
                    Join
                </button>
            </div>
        </div>
    </div>
</template>

<script>
    import M from 'materialize-css'
    import axios from 'axios'
    import {Config} from "../config"
    import $ from 'jquery'
    import 'materialize-css/dist/css/materialize.css'   

    export default {
        name: 'Login',
        data: function() {
            return  {
                username: "",
            }
        },

        methods: {
            join: function() {
                if (!this.username) {
                    M.toast({ html: 'You must choose a username', displayLength: 2000} );
                    return
                }
                this.login()
                this.username = $('<p>').html(this.username).text();
            },

            login: function() {
                let self = this
                axios.post(Config.HttpHost +  "login", this.username)
                .then(response => {
                    if (response.data.success) {
                        self.$store.commit("updateUsername", self.username)
                        this.$router.push({ name: 'Chatroom'})       
                        // window.location.href will cause a full page load and won't work here as Vuex state is stored in memory
                        // and we will lose this state on such page load.          
                    }
                    else{
                        M.toast({ html: 'Username already exists', displayLength: 2000} );
                        return
                    }
                })
                .catch(() => {
                    M.toast({ html: "Error!", displayLength: 2000} );
                    return
                })
            }
        }
    }
</script>
