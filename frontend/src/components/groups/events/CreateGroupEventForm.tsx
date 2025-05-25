"use client";

import React, { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { useRequest } from "@/hooks/useRequest";
import { GroupEvent } from "@/types/GroupEvent";
import {
  Dialog,
  DialogTitle,
  DialogBody,
  DialogActions,
} from "@/components/ui/dialog";
import { Button as HeadlessButton } from "@headlessui/react"; // For Dialog close

interface CreateGroupEventFormProps {
  groupId: string;
  onEventCreated: (event: GroupEvent) => void;
}

export function CreateGroupEventForm({
  groupId,
  onEventCreated,
}: CreateGroupEventFormProps) {
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [eventTime, setEventTime] = useState("");
  const [isOpen, setIsOpen] = useState(false);

  const { post, error: apiError, isLoading } = useRequest<GroupEvent>();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title || !eventTime) {
      alert("Title and Event Time are required.");
      return;
    }
    const payload = {
      title,
      description,
      event_time: new Date(eventTime).toISOString(),
    };
    const createdEvent = await post(
      `/api/groups/${groupId}/events`,
      payload,
      (apiResponseEvent: GroupEvent) => {
        const newEventForState = {
          ...apiResponseEvent,
        };
        onEventCreated(newEventForState);
        setIsOpen(false);
        setTitle("");
        setDescription("");
        setEventTime("");
      }
    );
    // Optionally handle createdEvent if needed, e.g. for further immediate actions
  };

  return (
    <>
      <Button onClick={() => setIsOpen(true)}>Create Event</Button>
      <Dialog open={isOpen} onClose={() => setIsOpen(false)}>
        <DialogTitle>Create New Event</DialogTitle>
        <DialogBody>
          <form onSubmit={handleSubmit} className="space-y-4 mt-4">
            <div>
              <label
                htmlFor="title"
                className="block text-sm font-medium text-gray-700"
              >
                Event Title
              </label>
              <Input
                id="title"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                required
              />
            </div>
            <div>
              <label
                htmlFor="description"
                className="block text-sm font-medium text-gray-700"
              >
                Description
              </label>
              <Textarea
                id="description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
              />
            </div>
            <div>
              <label
                htmlFor="eventTime"
                className="block text-sm font-medium text-gray-700"
              >
                Day/Time
              </label>
              <Input
                id="eventTime"
                type="datetime-local"
                value={eventTime}
                onChange={(e) => setEventTime(e.target.value)}
                required
              />
            </div>
            {apiError && (
              <div className="text-red-500 text-sm">
                <h4>Oops....</h4>
                <p>{apiError.message}</p>
              </div>
            )}
          </form>
        </DialogBody>
        <DialogActions>
          <HeadlessButton
            className="rounded bg-gray-200 py-2 px-4 text-sm text-gray-700 hover:bg-gray-300"
            onClick={() => setIsOpen(false)}
          >
            Cancel
          </HeadlessButton>
          <Button type="submit" onClick={handleSubmit} disabled={isLoading}>
            {isLoading ? "Creating..." : "Create Event"}
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
}
