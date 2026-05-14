<template>
  <div class="auth-success">
    <el-result icon="success" title="授权成功" sub-title="您的小程序已成功授权给附件通平台">
      <template #extra>
        <el-button type="primary" @click="$router.push('/dashboard')">进入后台</el-button>
      </template>
    </el-result>
  </div>
</template>

<style scoped>
.auth-success {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
