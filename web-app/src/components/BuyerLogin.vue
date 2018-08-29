
<template>
  <div>
    <input v-model="email" placeholder="edit me">
    <p>Email is: {{ email }}</p>
    <input v-model="password" placeholder="edit me">
    <p>Password is: {{ password }}</p>
    <button @click="submitLogin">Log In</button>
  </div>
</template>

<script>

export default {
  name: 'BuyerLogin',
  props: {
    msg: String
  },
  data() {
    return {
      email: '',
      password: '',
    }
  },
  methods: {
    submitLogin() {
      console.log("email: " + this.email);
      console.log("password: " + this.password);


      const buyerLoginRequest = {
        email: this.email,
        password: this.password
      };

      console.log(buyerLoginRequest);

      this.$store.dispatch('buyerLogin', buyerLoginRequest);
    } 
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
