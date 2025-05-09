module.exports = {
  apps: [
    {
      name: "minio",
      script: "minio",
      args: "server ./minio-data --address :3334",
      env_file: ".env",
      env: {
        MINIO_ROOT_USER: process.env.MINIO_ACCESS_KEY,
        MINIO_ROOT_PASSWORD: process.env.MINIO_SECRET_KEY,
      },
    },
    {
      name: "minio-ui",
      script: "./go-server",
      interpreter: "none",
      env_file: ".env",
    },
  ],
};
