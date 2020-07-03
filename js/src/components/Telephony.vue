<template>
  <div class="telephony mt-4">
    <!-- <div v-for="(chat, index) in chats" :key="index">
      <div class="card mb-0" style="height: 40px">
        <span class="ml-2">{{chat.Friend.username}}</span>
      </div>
    </div> -->
    <ul class="nav nav-pills flex-column">
      <div v-for="(chat, index) in chats" :key="index">
        <li :class="[`nav-item card mb-0 mt-0 conv-box-${index}`]" :data-index="index" ref="contact" style="height: 50px">
          <div class="row ml-2 mt-2" @click="renderChat" style="cursor: pointer">
            <i class="material-icons">person</i>
            <h4 class="ml-2 lead">{{chat.friend.username}}</h4>
          </div>
        </li>
      </div>
    </ul>
  </div>
</template>

<script>
  // import {Config} from "../config"

  export default {
    name: 'Telephony',
    props: {
      chats: {
        type: Array
      }
    },
    data: function() {
      return {
        selectedChat: 0,
      }
    },
    mounted: function() {
      if (this.$refs.contact != undefined)
        this.$refs.contact[this.selectedChat].style.background  = "#DFEEE9"
    },

    methods: {
      renderChat: function(e) {
        let card = e.currentTarget.closest(".card")
        this.$refs.contact.forEach(li => {
          li.style.background = "#fff"
        });
        card.style.background = "#DFEEE9"
        let index = card.getAttribute("data-index")
        this.selectedChat = index
        this.$emit('set-chat', index)
      }
    }
  }
</script>

<style>
</style>