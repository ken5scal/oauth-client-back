<template>
  <div class="hello">
    <h1>{{ msg }}</h1>
    <h2>Essential Links</h2>
    <button @click="apiAuthZ">Get Authorize</button>
  </div>
</template>

<script>
import axios from 'axios'
axios.defaults.headers.post['Content-Type'] = 'application/x-www-form-urlencoded'
export default {
  name: 'Authorize',
  data () {
    return {
      msg: 'hogefuga'
    }
  },
  methods: {
    apiAuthZ: async function () {
      let res = await axios.get('http://localhost:9000/authorize')
      this.msg = res.data
    }
  }
}
</script>
