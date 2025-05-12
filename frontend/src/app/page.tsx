'use client';

import { Heading } from '@/components/ui/heading';
import { Text } from '@/components/ui/text';
import { Button } from '@/components/ui/button';
import Link from 'next/link';
import { useRequest } from '@/hooks/useRequest';
import { useEffect } from 'react';
import { UserType } from '@/types/User';

export default function HomePage() {
  const { isLoading, data, error, get } = useRequest<UserType>();

  useEffect(() => {
    get('/api/users', (data) => {
      console.log(data);
    });
  }, []);

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
      {/* Enhanced Hero Section */}
      <div className="text-center mb-24 space-y-8">
        <div className="relative inline-block">
          <div className="absolute -inset-1 bg-gradient-to-r from-purple-600 to-pink-600 blur-2xl opacity-30 rounded-full"></div>
          <Heading
            level={1}
            className="relative text-6xl md:text-7xl font-bold mb-6 text-gray-900 bg-gradient-to-r from-indigo-500 via-purple-500 to-pink-500 text-transparent bg-clip-text"
          >
            Connect in Your Own Space
          </Heading>
        </div>

        <Text className="text-xl md:text-2xl text-gray-600 max-w-4xl mx-auto leading-relaxed">
          Join a social platform that values{' '}
          <span className="font-semibold text-purple-600">
            authentic connections
          </span>{' '}
          and{' '}
          <span className="font-semibold text-indigo-600">
            privacy-first design
          </span>
          . Share your world, discover communities, and engage in meaningful
          conversations.
        </Text>

        <div className="flex flex-col sm:flex-row justify-center items-center gap-6 mt-12">
          <Button
            href="/register"
            className="px-8 py-4 text-xl bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-700 hover:to-purple-700 transform transition-all hover:scale-105"
          >
            Start Your Journey →
          </Button>
          <Button
            href="/login"
            outline
            className="px-8 py-4 text-xl border-2 border-gray-300 hover:border-purple-500 hover:text-purple-600"
          >
            Existing Member? Sign In
          </Button>
        </div>

        {/* Hero Visual Element with Features */}
        <div className="mt-16 relative max-w-4xl mx-auto bg-gradient-to-br from-indigo-50 to-pink-50 rounded-2xl p-8 shadow-2xl group">
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-0 -mx-4 relative">
            {features.map((feature, index) => (
              <div
                key={index}
                className={`p-6 rounded-xl border border-gray-200 hover:border-purple-500 bg-white mx-4 transform transition-all duration-300 hover:scale-105 hover:z-10 group-hover:scale-100 group-hover:rotate-0 ${
                  index % 3 === 0
                    ? 'scale-90'
                    : index % 3 === 1
                    ? 'scale-95'
                    : 'scale-100'
                } ${
                  index % 2 === 0
                    ? 'rotate-2 origin-top'
                    : '-rotate-2 origin-bottom'
                }`}
              >
                <div className="w-12 h-12 flex items-center justify-center rounded-full bg-gradient-to-br from-indigo-100 to-purple-100 mb-4">
                  {feature.icon}
                </div>
                <h3 className="mb-3 text-xl font-semibold text-gray-900">
                  {feature.title}
                </h3>
                <Text className="text-gray-600">{feature.description}</Text>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Community Benefits Section */}
      <div className="text-center mb-16">
        <Heading level={2} className="text-4xl font-bold mb-8 text-gray-900">
          Why People Love Our Platform
        </Heading>
        <div className="grid md:grid-cols-3 gap-8">
          <div className="p-8 rounded-xl bg-gradient-to-br from-indigo-50 to-purple-50 hover:shadow-lg transition-all">
            <div className="w-16 h-16 flex items-center justify-center rounded-full bg-purple-600 mx-auto mb-4">
              <svg
                className="w-8 h-8 text-white"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
            </div>
            <h3 className="text-xl font-semibold mb-2 text-gray-900">
              Privacy First
            </h3>
            <Text className="text-gray-600">
              Full control over your data and visibility settings
            </Text>
          </div>

          <div className="p-8 rounded-xl bg-gradient-to-br from-purple-50 to-pink-50 hover:shadow-lg transition-all">
            <div className="w-16 h-16 flex items-center justify-center rounded-full bg-purple-600 mx-auto mb-4">
              <svg
                className="w-8 h-8 text-white"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
                />
              </svg>
            </div>
            <h3 className="text-xl font-semibold mb-2 text-gray-900">
              Vibrant Communities
            </h3>
            <Text className="text-gray-600">
              Join groups that match your interests and passions
            </Text>
          </div>

          <div className="p-8 rounded-xl bg-gradient-to-br from-pink-50 to-indigo-50 hover:shadow-lg transition-all">
            <div className="w-16 h-16 flex items-center justify-center rounded-full bg-purple-600 mx-auto mb-4">
              <svg
                className="w-8 h-8 text-white"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M13 10V3L4 14h7v7l9-11h-7z"
                />
              </svg>
            </div>
            <h3 className="text-xl font-semibold mb-2 text-gray-900">
              Instant Connection
            </h3>
            <Text className="text-gray-600">
              Real-time messaging and seamless communication
            </Text>
          </div>
        </div>
      </div>

      {/* CTA Section */}
      <div className="text-center bg-gradient-to-r from-indigo-100 via-purple-100 to-pink-100 rounded-2xl p-12">
        <h2 className="text-4xl font-bold mb-6 text-gray-900">
          Start Your Journey Today
        </h2>
        <Text className="text-xl mb-8 max-w-2xl mx-auto text-gray-700">
          Join our vibrant community and experience a social network that puts
          you in control. Create, connect, and communicate your way.
        </Text>
        <Button
          href="/register"
          className="px-6 py-3 text-lg bg-gradient-to-r from-indigo-500 to-purple-600 hover:from-indigo-600 hover:to-purple-700"
        >
          Create Your Account
        </Button>
      </div>
    </div>
  );
}

const features = [
  {
    title: 'Smart Following System',
    description:
      'Follow users and build your network with our intelligent request system that respects privacy preferences.',
    icon: (
      <svg
        className="w-6 h-6 text-purple-600"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"
        />
      </svg>
    ),
  },
  {
    title: 'Rich Profile System',
    description:
      'Express yourself with customizable profiles featuring avatars, bios, and activity feeds.',
    icon: (
      <svg
        className="w-6 h-6 text-purple-600"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M5.121 17.804A13.937 13.937 0 0112 16c2.5 0 4.847.655 6.879 1.804M15 10a3 3 0 11-6 0 3 3 0 016 0zm6 2a9 9 0 11-18 0 9 9 0 0118 0z"
        />
      </svg>
    ),
  },
  {
    title: 'Dynamic Posts',
    description:
      'Share moments with text, images, and GIFs. Control who sees your content with flexible privacy options.',
    icon: (
      <svg
        className="w-6 h-6 text-purple-600"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
        />
      </svg>
    ),
  },
  {
    title: 'Active Communities',
    description:
      'Create and join groups, organize events, and engage with like-minded people in group discussions.',
    icon: (
      <svg
        className="w-6 h-6 text-purple-600"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
        />
      </svg>
    ),
  },
  {
    title: 'Real-time Chat',
    description:
      'Connect instantly with private messaging, emoji support, and group chat rooms for seamless communication.',
    icon: (
      <svg
        className="w-6 h-6 text-purple-600"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
        />
      </svg>
    ),
  },
  {
    title: 'Smart Notifications',
    description:
      'Stay informed with real-time notifications for follows, group invites, and event updates.',
    icon: (
      <svg
        className="w-6 h-6 text-purple-600"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"
        />
      </svg>
    ),
  },
];
