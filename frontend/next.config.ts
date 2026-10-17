import type { NextConfig } from "next";

/** @type {import('next').NextConfig} */
const nextConfig = {
  devIndicators: false,
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: 'http://localhost:5001/api/:path*', // URL бэкенда на Go
      },
    ];
  },
};

export default nextConfig;

