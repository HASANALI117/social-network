/** @type {import('next').NextConfig} */
const nextConfig = {
  images: {
    remotePatterns: [
      {
        protocol: "http",
        hostname: "minio_local_storage",
        port: "9000", // As per the error message
        pathname: "/images/uploads/**",
      },
      {
        protocol: "http",
        hostname: "example.com",
      },
    ],
  },
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: "http://backend_app:8080/api/:path*",
      },
    ];
  },
};

export default nextConfig;
