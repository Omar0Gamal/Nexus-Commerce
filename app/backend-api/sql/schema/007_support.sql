-- ==================== Customer Support System ====================

CREATE TABLE support_tickets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID REFERENCES shops(id) ON DELETE CASCADE,

    -- The Customer (Reporter)
    customer_id UUID REFERENCES customers(id) ON DELETE SET NULL,
    guest_email VARCHAR(255),

    -- Ticket Details
    subject VARCHAR(255) NOT NULL,
    status ticket_status DEFAULT 'open',
    priority ticket_priority DEFAULT 'medium',
    category VARCHAR(50),

    -- Assignment
    assigned_staff_id UUID REFERENCES shop_staff(id) ON DELETE SET NULL,

    -- Links
    related_order_id UUID REFERENCES orders(id) ON DELETE SET NULL,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_tickets_shop_status ON support_tickets(shop_id, status);
CREATE INDEX idx_tickets_assigned_staff ON support_tickets(assigned_staff_id);

CREATE TABLE support_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID REFERENCES support_tickets(id) ON DELETE CASCADE,

    sender_type message_sender NOT NULL,

    -- If sender is Staff
    staff_id UUID REFERENCES shop_staff(id) ON DELETE SET NULL,

    -- If sender is Customer
    customer_id UUID REFERENCES customers(id) ON DELETE SET NULL,

    message_body TEXT NOT NULL,
    attachments JSONB DEFAULT '[]',

    is_internal_note BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP WITH TIME ZONE,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE support_automation_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID REFERENCES shops(id) ON DELETE CASCADE,

    name VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,

    -- Triggers (JSON Logic)
    conditions JSONB NOT NULL,

    -- Actions
    action_assign_staff_id UUID REFERENCES shop_staff(id),
    action_assign_priority ticket_priority,
    action_auto_reply_text TEXT,

    priority_order INTEGER DEFAULT 0
);
