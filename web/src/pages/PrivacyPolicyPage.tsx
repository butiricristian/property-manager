import { Link } from "react-router-dom";
import {
  Card,
  CardHeader,
  CardTitle,
  CardContent,
} from "@/components/ui/card";

export function PrivacyPolicyPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-4">
      <Card className="w-full max-w-2xl">
        <CardHeader>
          <CardTitle>Privacy Policy</CardTitle>
          <p className="text-sm text-muted-foreground">
            Last updated: April 12, 2026
          </p>
        </CardHeader>
        <CardContent className="prose prose-sm dark:prose-invert max-w-none space-y-6">
          <section>
            <h3 className="text-base font-semibold">1. Data We Collect</h3>
            <p className="text-sm text-muted-foreground leading-relaxed">
              We collect the following information when you use Property Manager:
            </p>
            <ul className="list-disc pl-5 text-sm text-muted-foreground space-y-1">
              <li>Email address and password (for email registration)</li>
              <li>Name, email, and profile picture (via Google OAuth)</li>
              <li>Property, booking, expense, and invoice data you enter</li>
              <li>Tax identification numbers included in billing profiles</li>
            </ul>
          </section>

          <section>
            <h3 className="text-base font-semibold">2. How We Use Your Data</h3>
            <p className="text-sm text-muted-foreground leading-relaxed">
              Your data is used solely to provide and improve the Property Manager
              service. We do not sell, rent, or share your personal data with third
              parties for marketing purposes.
            </p>
          </section>

          <section>
            <h3 className="text-base font-semibold">3. Data Storage & Security</h3>
            <p className="text-sm text-muted-foreground leading-relaxed">
              Data is stored in a PostgreSQL database. Passwords are hashed using
              bcrypt. All connections use HTTPS in production. We take reasonable
              measures to protect your data, but no system is 100% secure.
            </p>
          </section>

          <section>
            <h3 className="text-base font-semibold">4. Third-Party Services</h3>
            <p className="text-sm text-muted-foreground leading-relaxed">
              We use Google OAuth for optional sign-in. When you sign in with
              Google, we receive your name, email address, and profile picture. We
              do not access any other Google account data.
            </p>
          </section>

          <section>
            <h3 className="text-base font-semibold">5. Property Sharing</h3>
            <p className="text-sm text-muted-foreground leading-relaxed">
              You may share properties with other users. Sharing is user-initiated
              and grants read-only access to the shared property's data (bookings,
              expenses, and invoices). You can revoke shared access at any time.
            </p>
          </section>

          <section>
            <h3 className="text-base font-semibold">6. Data Retention</h3>
            <p className="text-sm text-muted-foreground leading-relaxed">
              Guest accounts and their data are automatically deleted after 7 days.
              Registered users can request account and data deletion at any time by
              contacting us.
            </p>
          </section>

          <section>
            <h3 className="text-base font-semibold">7. Your Rights (GDPR)</h3>
            <p className="text-sm text-muted-foreground leading-relaxed">
              If you are in the European Economic Area, you have the right to:
            </p>
            <ul className="list-disc pl-5 text-sm text-muted-foreground space-y-1">
              <li>Access the personal data we hold about you</li>
              <li>Request correction of inaccurate data</li>
              <li>Request deletion of your data</li>
              <li>Request a portable copy of your data</li>
              <li>Object to or restrict processing of your data</li>
            </ul>
          </section>

          <section>
            <h3 className="text-base font-semibold">8. Contact</h3>
            <p className="text-sm text-muted-foreground leading-relaxed">
              For privacy-related inquiries or to exercise your rights, contact us
              at{" "}
              <a
                href="mailto:privacy@propertymanager.app"
                className="text-primary hover:underline"
              >
                privacy@propertymanager.app
              </a>
              .
            </p>
          </section>

          <div className="pt-4 border-t">
            <Link to="/register" className="text-sm text-primary hover:underline">
              &larr; Back to registration
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
