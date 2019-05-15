package k8s

import (
	"bytes"
	"k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"math/rand"
	"testing"
)

const (
	testWebhookName = "c7c95710-d8c3-4cc3-a2a8-8d2b46909c76"
	testSecretName  = "15906410-af2a-4f9b-8a2d-c08ffdd5e129"
	testNamespace   = "7cad5f92-c0d5-4bc9-87a3-6f44d5a5619d"
)

func genSecretData() (ca []byte, cert []byte, key []byte) {
	ca = make([]byte, 4)
	cert = make([]byte, 4)
	key = make([]byte, 4)
	rand.Read(ca)
	rand.Read(cert)
	rand.Read(key)
	return
}

func newTestSimpleK8s() *k8s {
	c := k8s{}
	c.clientset = fake.NewSimpleClientset()
	return &c
}

func TestGetCaFromCertificate(t *testing.T) {
	client = newTestSimpleK8s()

	ca, cert, key := genSecretData()

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: testSecretName,
		},
		Data: map[string][]byte{"ca": ca, "cert": cert, "key": key},
	}

	client.clientset.CoreV1().Secrets(testNamespace).Create(secret)

	retrievedCa := GetCaFromCertificate(testSecretName, testNamespace)
	if !bytes.Equal(retrievedCa, ca) {
		t.Error("Was not able to retrieve CA information that was saved")
	}
}

func TestSaveCertsToSecret(t *testing.T) {
	client = newTestSimpleK8s()

	ca, cert, key := genSecretData()

	SaveCertsToSecret(testSecretName, testNamespace, ca, cert, key)

	secret, _ := client.clientset.CoreV1().Secrets(testNamespace).Get(testSecretName, metav1.GetOptions{})

	if !bytes.Equal(secret.Data["ca"], ca) {
		t.Error("'ca' saved data does not match retrieved")
	}

	if !bytes.Equal(secret.Data["cert"], cert) {
		t.Error("'cert' saved data does not match retrieved")
	}

	if !bytes.Equal(secret.Data["key"], key) {
		t.Error("'key' saved data does not match retrieved")
	}
}

func TestSaveThenLoadSecret(t *testing.T) {
	client = newTestSimpleK8s()
	ca, cert, key := genSecretData()
	SaveCertsToSecret(testSecretName, testNamespace, ca, cert, key)
	retrievedCa := GetCaFromCertificate(testSecretName, testNamespace)
	if !bytes.Equal(retrievedCa, ca) {
		t.Error("Was not able to retrieve CA information that was saved")
	}
}

func TestPatchWebhookConfigurations(t *testing.T) {
	client = newTestSimpleK8s()

	ca, _, _ := genSecretData()

	client.clientset.
		AdmissionregistrationV1beta1().
		MutatingWebhookConfigurations().
		Create(&v1beta1.MutatingWebhookConfiguration{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: testWebhookName,
			},
			Webhooks: []v1beta1.Webhook{{Name: "m1"}, {Name: "m2"}}})

	client.clientset.
		AdmissionregistrationV1beta1().
		ValidatingWebhookConfigurations().
		Create(&v1beta1.ValidatingWebhookConfiguration{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: testWebhookName,
			},
			Webhooks: []v1beta1.Webhook{{Name: "v1"}, {Name: "v2"}}})

	PatchWebhookConfigurations(testWebhookName, ca, true, true)

	whmut, err := client.clientset.
		AdmissionregistrationV1beta1().
		MutatingWebhookConfigurations().
		Get(testWebhookName, metav1.GetOptions{})

	whval, err := client.clientset.
		AdmissionregistrationV1beta1().
		MutatingWebhookConfigurations().
		Get(testWebhookName, metav1.GetOptions{})

	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(whmut.Webhooks[0].ClientConfig.CABundle, ca) {
		t.Error("Ca retrieved from webhook configuration does not match")
	}
	if !bytes.Equal(whmut.Webhooks[1].ClientConfig.CABundle, ca) {
		t.Error("Ca retrieved from webhook configuration does not match")
	}
	if !bytes.Equal(whval.Webhooks[0].ClientConfig.CABundle, ca) {
		t.Error("Ca retrieved from webhook configuration does not match")
	}
	if !bytes.Equal(whval.Webhooks[1].ClientConfig.CABundle, ca) {
		t.Error("Ca retrieved from webhook configuration does not match")
	}

}