<template>
  <div class="chat" >
      <div class="row min-vh-100 ml-4 mr-4">
          <div class="col-md-12">
              <div class="card horizontal h-100">
                  <div id="chat-messages" class="card-content" >
                    <div class="text-left" v-html="chatContent">
                    </div>
                  </div>
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
</template>

<script>
  import $ from 'jquery'
  import MD5 from 'crypto-js/md5'
  import {Config} from "../config"
  import 'materialize-css/dist/css/materialize.css'

  export default {
    name: 'Chatroom',
    props: {
      msg: String
    },
    computed: {
        currentUser: function() {
          return this.$store.getters.getUsername
        }
    },
    data: function() {
      return  {
        ws: null, // websocket
        newMessage: '',
        chatContent: '',
      }
    },

    created: function() {
        var self = this;
        this.ws = new WebSocket(Config.WsHost + 'ws');
        this.ws.addEventListener('message', function(e) {
            var msg = JSON.parse(e.data);
            self.chatContent += '<div class="chip">'
                    + '<img src="' + self.gravatarURL(msg.username) + '">'
                    + self.currentUser + '</div>'
                    + msg.message
                    + '<br/>';
            var element = document.getElementById('chat-messages');
            element.scrollTop = element.scrollHeight; // Auto scroll to the bottom
        });
    },

    methods: {
        send: function () {
            // debugger // eslint-disable-line
            if (this.newMessage != '') {
                this.ws.send(
                    JSON.stringify({
                          username: this.currentUser,
                          message: $('<p>').html(this.newMessage).text()
                    }
                ));
                this.newMessage = '';
            }
        },

        gravatarURL: function(email) {
            return 'http://www.gravatar.com/avatar/' + MD5(email);
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
