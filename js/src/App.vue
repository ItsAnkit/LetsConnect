<template>
  <div id="app">
    <header>
        <nav class="navbar navbar-light bg-light">
        <span>
            <a href="/" class="navbar-brand mb-0 h1">Let's Connect</a>
        </span>
        </nav>
    </header>
    <router-view></router-view>
  </div>
</template>

<script>
export default {
  name: 'app'
}
</script>