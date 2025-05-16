'use client';

import { useState, ReactNode, FC } from 'react';

interface Tab {
  label: string;
  content: ReactNode;
}

interface TabsProps {
  tabs: Tab[];
  initialTab?: number;
  onTabChange?: (index: number) => void; // Add onTabChange prop
}

interface PanelProps {
  id: string;
  className?: string;
  children: ReactNode;
}

// Panel component remains the same
const Panel: FC<PanelProps> = ({ children, ...props }) => {
  // The Panel component simply renders its children.
  // The id and className props are passed from GroupDetailPage,
  // and can be used here if needed for styling or accessibility.
  return <div {...props}>{children}</div>;
};

// The main Tabs functional component
const TabsFC: FC<TabsProps> = ({ tabs, initialTab = 0, onTabChange }) => { // Destructure onTabChange
  const [activeTab, setActiveTab] = useState(initialTab);

  const handleTabClick = (index: number) => {
    setActiveTab(index);
    if (onTabChange) {
      onTabChange(index);
    }
  };

  return (
    <div>
      <div className="border-b border-gray-200 dark:border-gray-700 mb-4">
        <nav className="-mb-px flex space-x-8" aria-label="Tabs">
          {tabs.map((tab, index) => (
            <button
              key={tab.label}
              onClick={() => handleTabClick(index)} // Call new handler
              className={`
                ${index === activeTab
                  ? 'border-indigo-500 text-indigo-600 dark:border-indigo-400 dark:text-indigo-300'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-200 dark:hover:border-gray-500'
                }
                whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm
              `}
              aria-current={index === activeTab ? 'page' : undefined}
            >
              {tab.label}
            </button>
          ))}
        </nav>
      </div>
      <div>
        {/* The content for the active tab is rendered here.
            The content itself is expected to use Tabs.Panel if it needs to.
            The Tabs component itself doesn't directly manage Panel instances,
            it just provides the structure for tab navigation and content display area.
            The actual <Tabs.Panel> usage is within the 'content' ReactNode of each tab.
        */}
        {tabs[activeTab] && tabs[activeTab].content}
      </div>
    </div>
  );
};

// Create a type that represents the Tabs component with a static Panel property
type TabsComponentType = FC<TabsProps> & {
  Panel: FC<PanelProps>;
};

// Cast TabsFC to this new type and assign Panel
const Tabs = TabsFC as TabsComponentType;
// Explicitly assign Panel to the exported Tabs constant.
// This ensures that Tabs.Panel is available and correctly typed.
Tabs.Panel = Panel;

export default Tabs;