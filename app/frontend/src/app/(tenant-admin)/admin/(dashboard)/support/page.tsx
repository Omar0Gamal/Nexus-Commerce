'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';

const toast = {
  success: (msg: string) => alert(msg),
  error: (msg: string) => alert(msg),
};

export default function SupportPage() {
  const [subject, setSubject] = useState('');
  const [message, setMessage] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    try {
      const res = await fetch('/api/v1/platform-tickets', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ subject, message }),
      });
      if (!res.ok) throw new Error('Failed to submit ticket');
      toast.success('Ticket submitted successfully');
      setSubject('');
      setMessage('');
    } catch (err) {
      toast.error('Failed to submit ticket');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="max-w-2xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-6">Platform Support</h1>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium mb-1">Subject</label>
          <Input 
            required 
            value={subject} 
            onChange={(e) => setSubject(e.target.value)} 
            placeholder="Issue summary"
          />
        </div>
        <div>
          <label className="block text-sm font-medium mb-1">Message</label>
          <Textarea 
            required 
            value={message} 
            onChange={(e) => setMessage(e.target.value)} 
            placeholder="Describe your issue in detail"
            rows={5}
          />
        </div>
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? 'Submitting...' : 'Submit Ticket'}
        </Button>
      </form>
    </div>
  );
}
