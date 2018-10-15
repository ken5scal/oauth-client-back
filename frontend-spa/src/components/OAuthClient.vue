<template>
  <div class="hello">
    <h1>{{ msg }}</h1>
    <h2>Essential Links</h2>
    <button @click="apiAuthZ">Get Authorize</button>
    <h2>Essential Links</h2>
    <ul>
      <li>
        <a
          href="https://vuejs.org"
          target="_blank"
        >
          Core Docs
        </a>
      </li>
      <li>
        <a
          href="https://forum.vuejs.org"
          target="_blank"
        >
          Forum
        </a>
      </li>
      <li>
        <a
          href="https://chat.vuejs.org"
          target="_blank"
        >
          Community Chat
        </a>
      </li>
      <li>
        <a
          href="https://twitter.com/vuejs"
          target="_blank"
        >
          Twitter
        </a>
      </li>
      <li>
        <a
          href="http://vuejs-templates.github.io/webpack/"
          target="_blank"
        >
          Docs for This Template
        </a>
      </li>
    </ul>
    <h2>Ecosystem</h2>
    <ul>
      <li>
        <a
          href="http://router.vuejs.org/"
          target="_blank"
        >
          vue-router
        </a>
      </li>
      <li>
        <a
          href="http://vuex.vuejs.org/"
          target="_blank"
        >
          vuex
        </a>
      </li>
      <li>
        <a
          href="http://vue-loader.vuejs.org/"
          target="_blank"
        >
          vue-loader
        </a>
      </li>
      <li>
        <a
          href="https://github.com/vuejs/awesome-vue"
          target="_blank"
        >
          awesome-vue
        </a>
      </li>
    </ul>
  </div>
</template>

<script>
import axios from 'axios'
axios.interceptors.response.use((response) => {
  console.log(response.headers)
  return response
}, (error) => {
  console.log('emergency')
  console.log(error)
  console.log(error.response)
  console.log(error.response.data)
  console.log(error.response.location)
  if (error.response && error.response.data && error.response.data.location) {
    window.location = error.response.data.location
  }
})

export default {
  name: 'Authorize',
  data () {
    return {
      msg: 'hogefuga'
    }
  },
  methods: {
    apiAuthZ: function () {
      fetch('http://localhost:9000/authorize', {
        // mode: 'cors',
        redirect: 'manual'
      }).then(response => {
        console.log(response)
        if (response.ok) {
          return response.blob()
        }

        if (response.type === 'opaqueredirect') {
          console.log(response)
          location.href = response.url
          return
        }

        throw new Error('network error')
      })
        .catch(error => {
          console.log('hoge', error.message)
        })
      // axios.get('http://localhost:9000/authorize').then(response => {
      //   console.log(response)
      //   console.log(response.headers)
      //   // window.location = 'https://login.microsoftonline.com'
      // })
    }
  }
}
</script>
