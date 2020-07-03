<template>
  <div class="lets-connect" >
    <div class="card-columns d-flex justify-content-center">
      <div class="card w-25 ml-4">
        <div class="card-block">
          <div class="row mb-0 ml-4">
            <div class="input-field w-75">
              <input type="number" v-model="searchContact" placeholder="  Search Contact" v-on:keyup="showSearchOption">
            </div>
            <div class="input-field ml-2">
              <button class="btn btn-success" type="submit" v-if="!searchContactPresent" @click="search">
                <i class="material-icons">search</i>
              </button>
              <button class="btn btn-success" type="submit" v-else @click="addContact">
                <i class="material-icons">person_add</i>
              </button>
            </div>
          </div>
          <span ref="searchErr" class="ml-4 mt-0"></span>
          <Telephony v-bind:chats="chats" v-on:set-chat="setChat"></Telephony>
        </div>
      </div>
      <div class="card w-75 ml-2 mr-4">
        <div class="card-block" v-if="chats.length">
          <div class="row min-vh-100">
            <div id="chat-messages" class="card-content" >
              <div class="text-left" v-html="chatContent">
              </div>
            </div>
          </div>
          <div class="row m-4">
            <div class="input-field w-75 ml-4">
              <input type="text" v-model="newMessage" @keyup.enter="send">
            </div>
            <div class="input-field col s4">
              <button class="waves-effect waves-light btn" @click="send">
                <i class="material-icons right">chat</i>
                  Send
              </button>
            </div>
          </div>
        </div>
        <div class="card-block min-vh-100" v-else>
          <h3 class="text-center">No chat History</h3>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
  import $ from 'jquery'
  import MD5 from 'crypto-js/md5'
  import {Config} from "../config"
  import axios from 'axios'
  import M from 'materialize-css'
  import Telephony from './Telephony.vue'
  import 'materialize-css/dist/css/materialize.css'

  export default {
    name: 'Home',
    components: {
      Telephony,
    },
    props: {
      msg: String
    },
    computed: {
        currentUser: function() {
          return this.$store.getters.getUser
        },
    },
    data: function() {
      return  {
        ws: null, // websocket
        newMessage: '',
        chatContent: '',
        conversations: '',
        searchContact: '',
        searchContactPresent: false,
        chats: [],
        currentChat: '',
      }
    },

    created: function() {
      var self = this;
      this.fetchChats();
      this.ws = new WebSocket(Config.WsHost + 'ws');
      this.ws.addEventListener('message', function(e) {
        console.log("jkl", e)
        var msg = JSON.parse(e.data);
        self.chatContent += '<div class="chip">'
                + '<img src="' + self.gravatarURL(msg.username) + '">'
                + self.currentUser.username + '</div>'
                + msg.message
                + '<br/>';
        var element = document.getElementById('chat-messages');
        element.scrollTop = element.scrollHeight;
      });
    },

    methods: {
      setChat: function(index) {
        this.currentChat = this.chats[index]
      },

      fetchChats: function() {
        let self = this
        axios.get(Config.HttpHost + `conversations/${self.currentUser.id}/chats`)
        .then( response => {
          if (response.data.success)
            self.chats = response.data.chats
            self.currentChat = response.data.chats[0]
        })
        .catch( () => {
          M.toast({ html: 'Some error occurred while fetching chats!!!', classes: 'black',  displayLength: 2000} );
          return
        })
      },

      send: function () {
        // debugger // eslint-disable-line
        if (this.newMessage != '') {
          this.ws.send(
            JSON.stringify({
              conversation_id: this.currentChat.conversation.id,
              sender_id: this.currentUser.id,
              message: $('<p>').html(this.newMessage).text()
            }
          ));
          this.newMessage = '';
        }
      },

      gravatarURL: function(email) {
        return 'http://www.gravatar.com/avatar/' + MD5(email);
      },

      search: function() {
        let self = this;
        if (self.searchContact == self.currentUser.mobile_no)  {
          $(self.$refs.searchErr).html("Sorry, You can't text yourself...")
          return
        }
        axios.get(Config.HttpHost + `conversations/${self.currentUser.id}/search?mobile_no=${self.searchContact}`)
        .then( response => {
          self.searchContactPresent = response.data.success
          if (response.data.success) {
            if (response.data.isContact) {
              self.searchContactPresent = false
              $(self.$refs.searchErr).html(response.data.message)
            } else 
              $(self.$refs.searchErr).html("You can add this person to contact list.")
          } else
            $(self.$refs.searchErr).html(response.data.message)
        })
        .catch( () => {
          M.toast({ html: 'Some error occurred while searching!!!', classes: 'black',  displayLength: 2000} );
          return
        })
      },

      addContact: function() {
        let self = this;
        axios.post(Config.HttpHost + `conversations/${self.currentUser.id}/add?friendMobile=${self.searchContact}`)
        .then((response) => {
          if (response.data.success) {
            self.chats = response.data.chats
          }
        })
        .catch(() => {
          M.toast({ html: 'Some error occurred while adding!!!', classes: 'black',  displayLength: 2000} );
          return
        })
      },

      showSearchOption: function() {
        this.searchContactPresent = false
        this.$refs.searchErr.innerHTML = ""
      }
    }
  }
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
  h3 {
    margin: 40px 0 0;
  }
  ul {
    list-style-type: none;
    padding: 0;
  }
  li {
    display: inline-block;
    margin: 0 10px;
  }
  a {
    color: #b97742;
  }

</style>
