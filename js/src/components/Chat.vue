<template>
  <div class="chat">
    <div class="card w-75 ml-2 mr-4">
      <div class="card-block">
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
    </div>
  </div>
</template>


<script>
  import {Config} from "../config"
  import $ from 'jquery'
  import MD5 from 'crypto-js/md5'
  // import M from 'materialize-css'

  export default {
    name: "Chat",
    props: ['conversation'],
    data: function() {
      return {
        ws: null,
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
      },
  },
}
</script>