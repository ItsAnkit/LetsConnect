<template>
    <div class="chat">
        <div class="row d-flex justify-content-center mt-4">
            <div class="input-field w-25">
                <input type="number" v-model.trim="mobile_no" placeholder="Mobile [0-9]" pattern="[0-9]{4}" required>
            </div>
        </div>
        <div class="row d-flex justify-content-center mt-4">
            <div class="input-field w-25">
                <input type="text" v-model.trim="username" placeholder="Username">
            </div>
        </div>
        <div class="row d-flex justify-content-center mt-4">
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
                mobile_no: "",
                chats: ""
            }
        },

        methods: {
            join: function() {
                // debugger // eslint-disable-line
                if (!this.mobile_no) {
                    M.toast({ html: 'You must provide a Mobile No.', classes: 'black',  displayLength: 2000} );
                    return
                }
                if (!this.username) {
                    M.toast({ html: 'You must provide a Username', classes: 'black', displayLength: 2000} );
                    return
                }
                this.login();
                this.username = $('<p>').html(this.username).text();
            },

            login: function() {
                let self = this
                axios.post(Config.HttpHost +  `login`, self.userDetails(), {
                    headers: { 'Content-Type': 'application/json' }
                })
                .then(response => {
                    if (response.data.success) {
                        self.$store.commit("updateUser", response.data.user)
                        this.$router.push({ name: 'Home'})
                        // window.location.href will cause a full page load and won't work here as Vuex state is stored in memory
                        // and we will lose this state on such page load.       
                    }
                    else{
                        M.toast({ html: 'Mobile Number already exists. Try different one.', classes: "black", displayLength: 2000} );
                        return
                    }
                })
                .catch((error) => {
                    M.toast({ html: error, classes: "black", displayLength: 2000} );
                    return
                })
            },
            userDetails: function() {
                return JSON.stringify({ 
                                            "mobile_no": this.mobile_no,
                                            "username": this.username,
                                      });
            }
        }
    }
</script>
