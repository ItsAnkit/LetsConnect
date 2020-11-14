<template>
  <div>
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
</template>


<script>
  import {Config} from "../config"
  import $ from 'jquery'
  import MD5 from 'crypto-js/md5'

  export default {
    name: "Chat",
    props: ['currentChat'],
    computed: {
        currentUser: function() {
          return this.$store.getters.getUser
        },
    },
    data: function() {
      return {
        ws: null,
        newMessage: '',
        chatContent: '',
      }
    },

    created: function() {
      var self = this;
      //var endPoint = `ws/users/${self.currentUser.id}/conversations/${self.currentChat.conversation.id}/ping`
      this.ws = new WebSocket(Config.WsHost + 'ws');
      this.ws.addEventListener('message', function(e) {
        var msg = JSON.parse(e.data);
        console.log("por", msg)
        self.chatContent += self.appendChat(msg, self.currentChat.friend)
        var element = document.getElementById('chat-messages');
        element.scrollTop = element.scrollHeight; // Auto scroll to the bottom
      });
    },

    mounted: function() {
      // this.send(true)
      this.ws.send(
        JSON.stringify({
          conversation_id: 0,
          sender_id: this.currentUser.id,
          message: ''
        }
      ));
    },

    methods: {
      appendChat: function(msg, user) {
        let chat = '<div class="chip">'
                + '<img src="' + 'http://www.gravatar.com/avatar/' + MD5(user.username) + '">'
                + user.username + '</div>'
                + msg
                + '<br/>';
        return chat
      },
    
      send: function (isCreated = false) {        
        //debugger // eslint-disable-line
        let message = $('<p>').html(this.newMessage).text()
        if (this.newMessage != '' || isCreated) {
          this.ws.send(
            JSON.stringify({
              conversation_id: this.currentChat.conversation.id,
              sender_id: this.currentUser.id,
              message: message
            }
          ));
          this.chatContent += this.appendChat(message, this.currentUser)
          this.newMessage = '';
        }
      },
  },
}
</script>