<script>
export default {
  name: 'pl-button',
  props: {
    autoLoading: {
      type: Boolean,
      default: true,
    },
  },
  data() {
    return {
      loadingStatus: false,
    }
  },
  methods: {
    handleClick() {
      if (this.autoLoading) {
        this.loadingStatus = true
      }
      this.$emit('click', () => {
        this.loadingStatus = false
      })
    },
  },
}
</script>

<el-button>
<slot/>
</el-button>
