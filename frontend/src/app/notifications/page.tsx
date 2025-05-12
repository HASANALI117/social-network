import ManageFollowRequestsSection from '@/components/notifications/ManageFollowRequestsSection';

export default function NotificationsPage() {
  return (
    <div className="p-4 sm:p-6 bg-gray-900 min-h-screen text-gray-100">
      <header className="mb-6">
        <h1 className="text-3xl font-bold text-white">Notifications</h1>
      </header>
      
      {/* Placeholder for other types of notifications if any */}
      {/* <div className="mb-8">
        <h2 className="text-xl font-semibold text-gray-200 mb-3">General Notifications</h2>
        <p className="text-gray-400">No new general notifications.</p>
      </div> */}

      <div>
        <h2 className="text-xl font-semibold text-gray-200 mb-3">Follow Requests</h2>
        <ManageFollowRequestsSection />
      </div>
    </div>
  );
}
