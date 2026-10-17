import { Link } from "react-router-dom";
import {
  Card,
  CardHeader,
  CardTitle,
  CardContent,
} from "@/components/ui/card";

export function TermsPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-4">
      <Card className="w-full max-w-2xl">
        <CardHeader>
          <CardTitle>Terms of Service</CardTitle>
          <p className="text-sm text-muted-foreground">
            Last updated: April 12, 2026
          </p>
        </CardHeader>
        <CardContent className="prose prose-sm dark:prose-invert max-w-none space-y-6">
          <section>
            <h3 className="text-base font-semibold">1. Acceptance of Terms</h3>
            <p className="text-sm text-muted-foreground leading-relaxed">
              By registering for or using Property Manager, you agree to be bound
              by these Terms of Service. If you do not agree, do not use the
              service.
            </p>
          </section>

          <section>
            <h3 className="text-base font-semibold">2. Account Types</h3>
            <p className="text-sm text-muted-foreground leading-relaxed">
              <strong>Guest accounts</strong> are temporary, expire after 7 days,
              and are limited to 1 property. Guest data is automatically deleted
              upon expiry.{" "}
              <strong>Registered accounts</strong> have full access to all features
              and no property limits.
            </p>
          </section>

          <section>
            <h3 className="text-base font-semibold">3. Service Description</h3>
            <p className="text-sm text-muted-foreground leading-relaxed">
              Property Manager is a dashboard for managing rental properties,
              bookings, expenses, invoices, and inventory. The service is provided
              for informational and organizational purposes.
            </p>
          </section>

          <section>
            <h3 className="text-base font-semibold">4. Financial Disclaimer</h3>
            <p className="text-sm text-muted-foreground leading-relaxed">
              Tax calculations, estimates, and summaries provided by Property
              Manager are for informational purposes only and do not constitute
              professional tax, legal, or financial advice. You are responsible for
              verifying all figures with a qualified professional before making
              financial decisions or filing tax returns.
            </p>
          </section>

          <section>
            <h3 className="text-base font-semibold">5. Property Sharing</h3>
            <p className="text-sm text-muted-foreground leading-relaxed">
              Property owners control who has access to their property data. Shared
              users receive read-only access. Owners can revoke access at any time.
              We are not responsible for any data shared between users.
            </p>
          </section>

          <section>
            <h3 className="text-base font-semibold">6. User Responsibilities</h3>
            <p className="text-sm text-muted-foreground leading-relaxed">
              You are responsible for maintaining the accuracy of data you enter,
              keeping your account credentials secure, and complying with
              applicable laws. You must not use the service for any unlawful
              purpose.
            </p>
          </section>

          <section>
            <h3 className="text-base font-semibold">7. Limitation of Liability</h3>
            <p className="text-sm text-muted-foreground leading-relaxed">
              The service is provided "as is" without warranties of any kind,
              express or implied. To the fullest extent permitted by law, we shall
              not be liable for any indirect, incidental, or consequential damages
              arising from your use of the service.
            </p>
          </section>

          <section>
            <h3 className="text-base font-semibold">8. Termination</h3>
            <p className="text-sm text-muted-foreground leading-relaxed">
              Either party may terminate at any time. You may delete your account
              by contacting us. We may suspend or terminate accounts that violate
              these terms. Upon termination, your data will be permanently deleted.
            </p>
          </section>

          <section>
            <h3 className="text-base font-semibold">9. Changes to Terms</h3>
            <p className="text-sm text-muted-foreground leading-relaxed">
              We may update these terms from time to time. Continued use of the
              service after changes constitutes acceptance of the updated terms. We
              will indicate the date of the latest revision at the top of this
              page.
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
