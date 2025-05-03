import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  /* config options here */
  output: "export",
  env: {
    API_URL: process.env.API_URL,
  },
};

export default nextConfig;
