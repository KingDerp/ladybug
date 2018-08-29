
<template>
  <div>
    <input v-model="email" placeholder="edit me">
    <p>Email is: {{ email }}</p>
    <input v-model="password" placeholder="edit me">
    <p>Password is: {{ password }}</p>
  </div>
</template>

<script>

export default {
  name: 'BuyerSingIn',
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
    submitSignIn() {
      console.log("email: " + this.email);
      console.log("password: " + this.password);


      const buyerSignInRequest = {
        email: this.email,
        password: this.password
      };

      console.log(buyerSignInRequest);

      this.$store.dispatch('buyerSignIn', buyerSignInRequest);
    } 
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
