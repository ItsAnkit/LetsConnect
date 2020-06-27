<template>
  <div class="telephony-list mt-4">
    <!-- <div v-for="(chat, index) in chats" :key="index">
      <div class="card mb-0" style="height: 40px">
        <span class="ml-2">{{chat.Friend.username}}</span>
      </div>
    </div> -->
    <ul class="nav nav-pills flex-column">
      <div v-for="(chat, index) in chats" :key="index">
        <li :class="[`nav-item card mb-0 mt-0 conv-box-${index}`]" :data-index="index" style="height: 50px">
          <div class="row ml-2 mt-2" @click="renderChat" style="cursor: pointer">
            <i class="material-icons">person</i>
            <h4 class="ml-2 lead">{{chat.Friend.username}}</h4>
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

    methods: {
      renderChat: function(e) {
        let index = e.currentTarget.closest(".card").getAttribute("data-index")
        this.$emit('set-chat', index)
      }
    }
  }
</script>

<style>
</style>